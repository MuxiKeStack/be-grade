package grpc

import (
	"context"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	gradev1 "github.com/MuxiKeStack/be-api/gen/proto/grade/v1"
	"github.com/MuxiKeStack/be-grade/domain"
	"github.com/MuxiKeStack/be-grade/service"
	"github.com/ecodeclub/ekit/slice"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

type GradeServiceServer struct {
	gradev1.UnimplementedGradeServiceServer
	svc service.GradeService
}

func NewGradeServiceServer(svc service.GradeService) *GradeServiceServer {
	return &GradeServiceServer{svc: svc}
}

func (g *GradeServiceServer) Register(server *grpc.Server) {
	gradev1.RegisterGradeServiceServer(server, g)
}

func (g *GradeServiceServer) GetGradesByCourseId(ctx context.Context, request *gradev1.GetGradesByCourseIdRequest) (*gradev1.GetGradesByCourseIdResponse, error) {
	grades, err := g.svc.GetGradesByCourseId(ctx, request.GetCourseId())
	return &gradev1.GetGradesByCourseIdResponse{
		Grades: slice.Map(grades, func(idx int, src domain.Grade) *ccnuv1.Grade {
			return &ccnuv1.Grade{
				Regular: src.Regular,
				Final:   src.Final,
				Total:   src.Total,
				Year:    src.Year,
				Term:    src.Term,
			}
		}),
	}, err
}

// todo 暴露一个重复签约的错误给上游
func (g *GradeServiceServer) SignForGradeSharing(ctx context.Context, request *gradev1.SignForGradeSharingRequest) (*gradev1.SignForGradeSharingResponse, error) {
	err := g.svc.SignForGradeSharing(ctx, request.GetUid(), request.GetWantsToSign())
	switch {
	case err == service.ErrRepeatSigned:
		return &gradev1.SignForGradeSharingResponse{}, gradev1.ErrorRepeatSigning("重复签约")
	case err == service.ErrRepeatCancelSigned:
		return &gradev1.SignForGradeSharingResponse{}, gradev1.ErrorRepeatCancelSigning("重复取消签约")
	default:
		return &gradev1.SignForGradeSharingResponse{}, err
	}
}

func (g *GradeServiceServer) ShareGrade(ctx context.Context, request *gradev1.ShareGradeRequest) (*gradev1.ShareGradeResponse, error) {
	err := g.svc.ShareGrade(ctx, request.GetUid(), request.GetStudentId(), request.GetPassword())
	if err == service.ErrNotSigned {
		return &gradev1.ShareGradeResponse{}, gradev1.ErrorNotSigned("尚未签约")
	}
	return &gradev1.ShareGradeResponse{}, err
}

func (g *GradeServiceServer) GetSignStatus(ctx context.Context, request *gradev1.GetSignStatusRequest) (*gradev1.GetSignStatusResponse, error) {
	isSigned, err := g.svc.IsSigned(ctx, request.GetUid())
	return &gradev1.GetSignStatusResponse{
		IsSigned: isSigned,
	}, err
}
