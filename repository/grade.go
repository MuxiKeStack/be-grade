package repository

import (
	"context"
	"github.com/MuxiKeStack/be-grade/domain"
	// repository/dao 这部分不区分大小写
	"github.com/MuxiKeStack/be-grade/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

var ErrRepeatSignOperation = dao.ErrRepeatSignOperation

type GradeRepository interface {
	GetGradesByCourseId(ctx context.Context, courseId int64) ([]domain.Grade, error)
	SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error
	UpsertGrades(ctx context.Context, grades []domain.Grade) error
	IsSigned(ctx context.Context, uid int64) (bool, error)
}

type gradeRepository struct {
	dao dao.GradeDAO
}

func NewGradeRepository(dao dao.GradeDAO) GradeRepository {
	return &gradeRepository{dao: dao}
}

func (repo *gradeRepository) GetGradesByCourseId(ctx context.Context, courseId int64) ([]domain.Grade, error) {
	courses, err := repo.dao.GetGradesByCourseId(ctx, courseId)
	return slice.Map(courses, func(idx int, src dao.Grade) domain.Grade {
		return repo.toDomain(src)
	}), err
}

func (repo *gradeRepository) SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error {
	return repo.dao.SignForGradeSharing(ctx, uid, wantsToSign)
}

func (repo *gradeRepository) UpsertGrades(ctx context.Context, grades []domain.Grade) error {
	return repo.dao.UpsertGrades(ctx, slice.Map(grades, func(idx int, src domain.Grade) dao.Grade {
		return repo.toEntity(src)
	}))
}

func (repo *gradeRepository) IsSigned(ctx context.Context, uid int64) (bool, error) {
	gs, err := repo.dao.GetGradeShareAgreements(ctx, uid)
	switch {
	case err == nil:
		return gs.IsSigned, nil
	case err == dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (repo *gradeRepository) toEntity(grade domain.Grade) dao.Grade {
	return dao.Grade{
		Id:       grade.Id,
		CourseId: grade.CourseId,
		Uid:      grade.Uid,
		Regular:  grade.Regular,
		Final:    grade.Final,
		Total:    grade.Total,
		Year:     grade.Year,
		Term:     grade.Term,
	}
}

func (repo *gradeRepository) toDomain(grade dao.Grade) domain.Grade {
	return domain.Grade{
		Id:       grade.Id,
		CourseId: grade.CourseId,
		Uid:      grade.Uid,
		Regular:  grade.Regular,
		Final:    grade.Final,
		Total:    grade.Total,
		Year:     grade.Year,
		Term:     grade.Term,
	}
}
