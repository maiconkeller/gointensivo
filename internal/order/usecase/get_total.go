package usecase

import "github.com/maiconkeller/gointensivo/internal/order/entity"

type GetTotalOutputDTO struct {
	Total int
}

type GetTotalUseCase struct{
	OrderRepository entity.OrderRepositoryInterface
}

func NewGetTotalUseCase(orderRepository entity.OrderRepositoryInterface) GetTotalUseCase {
	return GetTotalUseCase{
        OrderRepository: orderRepository,
    }
}

func (u *GetTotalUseCase) Execute() (*GetTotalOutputDTO, error) {
	total, err := u.OrderRepository.GetTotal()
    if err!= nil {
        return nil, err
    }
	return &GetTotalOutputDTO{
        Total: total,
	}, nil
}