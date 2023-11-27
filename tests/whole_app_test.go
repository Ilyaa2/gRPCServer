package tests

import (
	mock_repository "gRPCServer/internal/repository/mocks"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestWholeServ(t *testing.T) {
	// инициализируем контроллер gomock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_repository.NewMockEmployee(ctrl)
	_ = mockRepo
	// mockRepo, ожидаем, что mockedVisitorLister будет вызван один раз с аргументом party.NiceVisitor и вернёт []string{“Peter”, "TheSmart"}, nil
	//mockRepo.EXPECT().ListVisitors(party.NiceVisitor).Return([]party.Visitor{{"Peter", "TheSmart"}}, nil)
	//mockRepo.EXPECT().GetByEmail
}
