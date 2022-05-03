package ephemeral_test

// func TestEphemeralQueue(t *testing.T) {
// 	sys := swarm.New("test")
// 	queue := swarm.Must(sut.New(sys, "test"))

// 	send, _ := queue.Send("ephemeral.test")
// 	recv, _ := queue.Recv("ephemeral.test")

// 	for i := 0; i < 100; i++ {
// 		send <- swarm.Bytes(fmt.Sprintf("%d", i))
// 	}

// 	for i := 0; i < 100; i++ {
// 		seen := <-recv
// 		it.Ok(t).
// 			If(seen).Equal(swarm.Bytes(fmt.Sprintf("%d", i)))
// 	}

// 	sys.Stop()
// }
