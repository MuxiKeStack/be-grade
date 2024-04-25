//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-grade/grpc"
	"github.com/MuxiKeStack/be-grade/ioc"
	"github.com/MuxiKeStack/be-grade/pkg/grpcx"
	"github.com/MuxiKeStack/be-grade/repository"
	"github.com/MuxiKeStack/be-grade/repository/dao"
	"github.com/MuxiKeStack/be-grade/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewGradeServiceServer,
		service.NewGradeService,
		repository.NewGradeRepository,
		dao.NewGORMGradeDAO,
		ioc.InitCCNUClient,
		ioc.InitCourseClient,
		ioc.InitDB,
		ioc.InitLogger,
		ioc.InitEtcdClient,
	)
	return grpcx.Server(nil)
}
