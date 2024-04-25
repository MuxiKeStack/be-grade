package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/MuxiKeStack/be-grade/pkg/logger"
	"github.com/MuxiKeStack/be-grade/pkg/saramax"
	"github.com/MuxiKeStack/be-grade/service"
	"time"
)

type ShareGradeEvent struct {
	Uid       int64
	StudentId string
	Password  string
}

type ShareGradeEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.GradeService
}

func NewShareGradeEventConsumer(client sarama.Client, l logger.Logger, svc service.GradeService) *ShareGradeEventConsumer {
	return &ShareGradeEventConsumer{client: client, l: l, svc: svc}
}

func (s *ShareGradeEventConsumer) Start() error {
	// 这里其实认为cg New成功了就算启动成功
	cg, err := sarama.NewConsumerGroupFromClient("update_shared_grade", s.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(),
			[]string{"share_grade_event"},
			saramax.NewHandler(s.l, s.Consume))
		if er != nil {
			s.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (s *ShareGradeEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ShareGradeEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return s.svc.ShareGrade(ctx, evt.Uid, evt.StudentId, evt.Password)
}
