package dao

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrRepeatSignOperation = errors.New("重复的签约操作")
	ErrRecordNotFound      = gorm.ErrRecordNotFound
)

type GradeDAO interface {
	GetGradesByCourseId(ctx context.Context, courseId int64) ([]Grade, error)
	SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error
	UpsertGrades(ctx context.Context, grades []Grade) error
	GetGradeShareAgreements(ctx context.Context, uid int64) (GradeShareAgreements, error)
}

type GORMGradeDAO struct {
	db *gorm.DB
}

func NewGORMGradeDAO(db *gorm.DB) GradeDAO {
	return &GORMGradeDAO{db: db}
}

func (dao *GORMGradeDAO) GetGradeShareAgreements(ctx context.Context, uid int64) (GradeShareAgreements, error) {
	var gs GradeShareAgreements
	err := dao.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&gs).Error
	return gs, err
}

func (dao *GORMGradeDAO) GetGradesByCourseId(ctx context.Context, courseId int64) ([]Grade, error) {
	var courses []Grade
	err := dao.db.WithContext(ctx).
		Where("course_id = ?", courseId).
		Find(&courses).Error
	return courses, err
}

func (dao *GORMGradeDAO) SignForGradeSharing(ctx context.Context, uid int64, wantsToSign bool) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var gs GradeShareAgreements
		err := tx.Where("uid = ?", uid).First(&gs).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		now := time.Now().UnixMilli()
		switch {
		case err == nil:
			if gs.IsSigned == wantsToSign {
				return ErrRepeatSignOperation
			}
			// 因为是检查，然后做某事，肯定有并发问题，但这里有问题!=有影响，况且一个用自己能有多少并发
			return tx.Model(&GradeShareAgreements{}).
				Where("uid = ?", uid).
				Updates(map[string]any{
					"utime":     now,
					"is_signed": wantsToSign,
				}).Error
		case err == gorm.ErrRecordNotFound:
			// 创建
			gs.Uid = uid
			gs.IsSigned = wantsToSign
			gs.Utime = now
			gs.Ctime = now
			return tx.Create(&gs).Error
		default:
			return err
		}
	})
}

func (dao *GORMGradeDAO) UpsertGrades(ctx context.Context, grades []Grade) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var eg errgroup.Group
		for _, g := range grades {
			eg.Go(func() error {
				return tx.Clauses(
					clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
						"utime": now,
					})}).Create(&g).Error
			})
		}
		return eg.Wait()
	})
}

type Grade struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	CourseId int64 `gorm:"uniqueIndex:cid_uid_year_term"` // 主键id
	Uid      int64 `gorm:"uniqueIndex:cid_uid_year_term"`
	Regular  float64
	Final    float64
	Total    float64
	Year     string `gorm:"uniqueIndex:cid_uid_year_term; type:char(4)"`
	Term     string `gorm:"uniqueIndex:cid_uid_year_term; type:char(1)"`
	Utime    int64
	Ctime    int64
}

type GradeShareAgreements struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Uid      int64 `gorm:"uniqueIndex"`
	IsSigned bool
	Utime    int64
	Ctime    int64
}
