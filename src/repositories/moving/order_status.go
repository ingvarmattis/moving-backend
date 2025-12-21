package moving

type OrderStatus int

const (
	OrderStatusUnknown OrderStatus = iota
	OrderStatusCreated
	OrderStatusRejected
	OrderStatusInProgress
	OrderStatusDone
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusCreated:
		return "created"
	case OrderStatusRejected:
		return "rejected"
	case OrderStatusInProgress:
		return "in_progress"
	case OrderStatusDone:
		return "done"
	default:
		return "unknown"
	}
}

func NewOrderStatus(s string) OrderStatus {
	switch s {
	case "created":
		return OrderStatusCreated
	case "rejected":
		return OrderStatusRejected
	case "in_progress":
		return OrderStatusInProgress
	case "done":
		return OrderStatusDone
	default:
		return OrderStatusUnknown
	}
}
