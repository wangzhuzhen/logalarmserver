package tchannel

import (
	"fmt"
	"github.com/uber-go/atomic"
	"context"
)

type HttpSystemError struct{
	msg string
	code SystemErrCode
	wrapped error
	nextMessageID atomic.Uint32
}

func (se HttpSystemError) Code() SystemErrCode {
	return se.code
}

func (se HttpSystemError) Message() string{
	return se.msg
}

func (se HttpSystemError) Wrapped() error {
	return se.wrapped
}


func (se HttpSystemError) Error() string{
	return fmt.Sprintf("system error %v : % s", se.Code(), se.Message())
}

func (se HttpSystemError) SendSystemError(ctx context.Context, port string, traceId uint64, spanId uint64, parentId uint64, flags byte, err error) {
	frame := NewFrame(MaxFramePayloadSize)
	id := se.nextMessageID.Inc()

	span := Span{
		traceID:traceId,
		spanID:spanId,
		parentID:parentId,
		flags:flags,
	}

	conn, _ := dialContext(ctx, port);
	frame.WriteOut(conn)

	if err := frame.write(&errorMessage{
		id : id,
		errCode: ErrCodeBadRequest,
		tracing: span,
		message: "Testxj123",
	}); err != nil{
		//return fmt.Errorf("failed to create outbound error frame")
	}



	//return nil
}