package service

import (
	"context"
	"errors"
	"fmt"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	"github.com/MuxiKeStack/be-grade/domain"
	"github.com/MuxiKeStack/be-grade/repository"
	"github.com/ecodeclub/ekit/slice"
)

var (
	ErrNotSigned          = errors.New("尚未签约")
	ErrRepeatSigned       = errors.New("重复签约")
	ErrRepeatCancelSigned = errors.New("重复取消签约")
)

type GradeService interface {
	GetGradesByCourseId(ctx context.Context, courseId int64) ([]domain.Grade, error)
	SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error
	ShareGrade(ctx context.Context, uid int64, studentId string, password string) error
	IsSigned(ctx context.Context, uid int64) (bool, error)
}

type gradeService struct {
	repo         repository.GradeRepository
	ccnuClient   ccnuv1.CCNUServiceClient
	courseClient coursev1.CourseServiceClient
}

func NewGradeService(repo repository.GradeRepository, ccnuClient ccnuv1.CCNUServiceClient, courseClient coursev1.CourseServiceClient) GradeService {
	return &gradeService{repo: repo, ccnuClient: ccnuClient, courseClient: courseClient}
}

func (g *gradeService) IsSigned(ctx context.Context, uid int64) (bool, error) {
	return g.repo.IsSigned(ctx, uid)
}

func (g *gradeService) GetGradesByCourseId(ctx context.Context, courseId int64) ([]domain.Grade, error) {
	return g.repo.GetGradesByCourseId(ctx, courseId)
}

func (g *gradeService) SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error {
	err := g.repo.SignForGradeSharing(ctx, uid, wantsToSign)
	if err == repository.ErrRepeatSignOperation {
		if wantsToSign {
			return ErrRepeatSigned
		} else {
			return ErrRepeatCancelSigned
		}
	}
	return err
}

// ShareGrade 慢的要死，5s，用kafka异步调
func (g *gradeService) ShareGrade(ctx context.Context, uid int64, studentId string, password string) error {
	// 0. 是否签约
	isSigned, err := g.IsSigned(ctx, uid)
	if err != nil {
		return err
	}
	if !isSigned {
		return ErrNotSigned
	}
	// 1. 爬取所有成绩
	res, err := g.ccnuClient.GetAllGrades(ctx, &ccnuv1.GetAllGradesRequest{
		StudentId: studentId,
		Password:  password,
	})
	if err != nil {
		return err
	}
	courseIdsRes, err := g.courseClient.FindIdsOrUpsertByCourses(ctx, &coursev1.FindIdOrUpsertByCoursesRequest{
		Courses: slice.Map(res.GetGrades(), func(idx int, src *ccnuv1.Grade) *coursev1.Course {
			return &coursev1.Course{
				CourseCode: src.GetCourseCode(),
				Name:       src.GetCourseName(),
				Teacher:    src.GetCourseTeacher(),
			}
		}),
	})
	if err != nil {
		return err
	}
	// 进行 course_code,teacher,name ==> courseId 的映射
	CourseIdsMap := slice.ToMap(courseIdsRes.GetCourses(), func(element *coursev1.Course) string {
		return g.courseIdKey(element.GetCourseCode(), element.GetName(), element.GetTeacher())
	})
	grades := slice.Map(res.GetGrades(), func(idx int, src *ccnuv1.Grade) domain.Grade {
		return domain.Grade{
			// 我要拿到courseId
			CourseId: CourseIdsMap[g.courseIdKey(src.GetCourseCode(), src.GetCourseName(), src.GetCourseTeacher())].GetId(),
			Uid:      uid,
			Regular:  src.Regular,
			Final:    src.Final,
			Total:    src.Total,
			Year:     src.Year,
			Term:     src.Term,
		}
	})
	// 所有的在一个事务里面upsert
	return g.repo.UpsertGrades(ctx, grades)
}

func (g *gradeService) courseIdKey(courseCode string, name string, teacher string) string {
	return fmt.Sprintf("%s-%s-%s", courseCode, name, teacher)
}
