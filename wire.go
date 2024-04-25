//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-grade/events"
	"github.com/MuxiKeStack/be-grade/grpc"
	"github.com/MuxiKeStack/be-grade/ioc"
	"github.com/MuxiKeStack/be-grade/repository"
	"github.com/MuxiKeStack/be-grade/repository/dao"
	"github.com/MuxiKeStack/be-grade/service"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		// consumer
		ioc.InitConsumers,
		events.NewShareGradeEventConsumer,
		ioc.InitKafka,
		// grpc
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
	return &App{}
}
