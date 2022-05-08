package system

import "time"

type Closer interface {
	Close()
}

/*

MsgSendCh is the pair of channel, exposed by the queue to clients to send messages
*/
type MsgSendCh[T any] struct {
	Msg chan T // channel to send message out
	Err chan T // channel to recv failed messages
}

func (ch *MsgSendCh[T]) Close() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Err) == 0 {
			break
		}
	}

	close(ch.Msg)
	close(ch.Err)
}

/*

msgRecv is the pair of channel, exposed by the queue to clients to recv messages
*/
type MsgRecvCh[T any] struct {
	Msg chan T // channel to recv message
	Ack chan T // channel to send acknowledgement
}

func (ch *MsgRecvCh[T]) Close() {
	for {
		time.Sleep(100 * time.Millisecond)
		if len(ch.Msg)+len(ch.Ack) == 0 {
			break
		}
	}

	close(ch.Msg)
	close(ch.Ack)
}
