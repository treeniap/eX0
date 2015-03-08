package main

import (
	"fmt"
	"net"
)

func ExampleFullConnection() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("panic:", err)
		}
	}()

	switch 2 {
	case 0:
		// Normal TCP + UDP.
		go server()
		client()
	case 1:
		// TCP and UDP via local channels. Requires `go test -tags=chan`.
		clientToServerTcp := make(chan []byte)
		serverToClientTcp := make(chan []byte)
		clientToServerUdp := make(chan []byte)
		serverToClientUdp := make(chan []byte)

		var clientToServerConn = &Connection{
			sendTcp: clientToServerTcp,
			recvTcp: serverToClientTcp,
			sendUdp: clientToServerUdp,
			recvUdp: serverToClientUdp,
		}

		var serverToClientConn = &Connection{
			sendTcp: serverToClientTcp,
			recvTcp: clientToServerTcp,
			sendUdp: serverToClientUdp,
			recvUdp: clientToServerUdp,
		}

		state.mu.Lock()
		state.connections = append(state.connections, serverToClientConn)
		state.mu.Unlock()

		go handleTcpConnection(serverToClientConn)
		go handleUdp(serverToClientConn)
		go sendServerUpdates()
		go broadcastPingPacket()
		fmt.Println("Started.")

		connectToServer(clientToServerConn)
	case 2:
		// Virtual TCP and UDP via physical TCP. Requires `go test -tags=tcp`.

		c := make(chan *Connection)
		go func(c chan *Connection) {
			var serverToClientConn = newConnection()
			ln, err := net.Listen("tcp", "localhost:25045")
			if err != nil {
				panic(err)
			}
			tcp, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			serverToClientConn.tcp = tcp
			close(serverToClientConn.start)
			c <- serverToClientConn
		}(c)

		s := make(chan *Connection)
		go func(s chan *Connection) {
			var clientToServerConn = newConnection()
			tcp, err := net.Dial("tcp", "localhost:25045")
			if err != nil {
				panic(err)
			}
			clientToServerConn.tcp = tcp
			close(clientToServerConn.start)
			s <- clientToServerConn
		}(s)

		var clientToServerConn = <-s
		var serverToClientConn = <-c

		state.mu.Lock()
		state.connections = append(state.connections, serverToClientConn)
		state.mu.Unlock()

		go handleTcpConnection(serverToClientConn)
		go handleUdp(serverToClientConn)
		go sendServerUpdates()
		go broadcastPingPacket()
		fmt.Println("Started.")

		connectToServer(clientToServerConn)
	}

	// Output:
	// Started.
	// (packet.JoinServerRequest)(packet.JoinServerRequest{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(26),
	// 		Type:   (packet.Type)(1),
	// 	}),
	// 	Version: (uint16)(1),
	// 	Passphrase: ([16]uint8)([16]uint8{
	// 		(uint8)(115),
	// 		(uint8)(111),
	// 		(uint8)(109),
	// 		(uint8)(101),
	// 		(uint8)(114),
	// 		(uint8)(97),
	// 		(uint8)(110),
	// 		(uint8)(100),
	// 		(uint8)(111),
	// 		(uint8)(109),
	// 		(uint8)(112),
	// 		(uint8)(97),
	// 		(uint8)(115),
	// 		(uint8)(115),
	// 		(uint8)(48),
	// 		(uint8)(49),
	// 	}),
	// 	Signature: (uint64)(123),
	// })
	// (packet.JoinServerAccept)(packet.JoinServerAccept{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(2),
	// 		Type:   (packet.Type)(2),
	// 	}),
	// 	YourPlayerId:     (uint8)(0),
	// 	TotalPlayerCount: (uint8)(15),
	// })
	// (packet.Handshake)(packet.Handshake{
	// 	UdpHeader: (packet.UdpHeader)(packet.UdpHeader{
	// 		Type: (packet.Type)(0),
	// 	}),
	// 	Signature: (uint64)(123),
	// })
	// (packet.UdpConnectionEstablished)(packet.UdpConnectionEstablished{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(0),
	// 		Type:   (packet.Type)(5),
	// 	}),
	// })
	// (packet.LocalPlayerInfo)(packet.LocalPlayerInfo{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(11),
	// 		Type:   (packet.Type)(30),
	// 	}),
	// 	NameLength: (uint8)(8),
	// 	Name: ([]uint8)([]uint8{
	// 		(uint8)(115),
	// 		(uint8)(104),
	// 		(uint8)(117),
	// 		(uint8)(114),
	// 		(uint8)(99),
	// 		(uint8)(111),
	// 		(uint8)(111),
	// 		(uint8)(76),
	// 	}),
	// 	CommandRate: (uint8)(20),
	// 	UpdateRate:  (uint8)(20),
	// })
	// (packet.LoadLevel)(packet.LoadLevel{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(5),
	// 		Type:   (packet.Type)(20),
	// 	}),
	// 	LevelFilename: ([]uint8)([]uint8{
	// 		(uint8)(116),
	// 		(uint8)(101),
	// 		(uint8)(115),
	// 		(uint8)(116),
	// 		(uint8)(51),
	// 	}),
	// })
	// (string)("test3")
	// (packet.CurrentPlayersInfo)(packet.CurrentPlayersInfo{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(25),
	// 		Type:   (packet.Type)(21),
	// 	}),
	// 	Players: ([]packet.PlayerInfo)([]packet.PlayerInfo{
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(8),
	// 			Name: ([]uint8)([]uint8{
	// 				(uint8)(115),
	// 				(uint8)(104),
	// 				(uint8)(117),
	// 				(uint8)(114),
	// 				(uint8)(99),
	// 				(uint8)(111),
	// 				(uint8)(111),
	// 				(uint8)(76),
	// 			}),
	// 			Team:  (uint8)(2),
	// 			State: (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 	}),
	// })
	// (packet.EnterGamePermission)(packet.EnterGamePermission{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(0),
	// 		Type:   (packet.Type)(6),
	// 	}),
	// })
	// (packet.EnteredGameNotification)(packet.EnteredGameNotification{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(0),
	// 		Type:   (packet.Type)(7),
	// 	}),
	// })
	// (packet.JoinTeamRequest)(packet.JoinTeamRequest{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(1),
	// 		Type:   (packet.Type)(27),
	// 	}),
	// 	PlayerNumber: (*uint8)(nil),
	// 	Team:         (uint8)(0),
	// })
	// (packet.PlayerJoinedTeam)(packet.PlayerJoinedTeam{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(15),
	// 		Type:   (packet.Type)(28),
	// 	}),
	// 	PlayerId: (uint8)(0),
	// 	Team:     (uint8)(0),
	// 	State: (*packet.State)(&packet.State{
	// 		CommandSequenceNumber: (uint8)(123),
	// 		X: (float32)(1),
	// 		Y: (float32)(2),
	// 		Z: (float32)(3),
	// 	}),
	// })
	// done
}

// Requires an empty server to be running.
func disabledExampleConnectToEmptyRealServer() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("panic:", err)
		}
	}()

	client()

	// Output:
	// (packet.JoinServerAccept)(packet.JoinServerAccept{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(2),
	// 		Type:   (packet.Type)(2),
	// 	}),
	// 	YourPlayerId:     (uint8)(0),
	// 	TotalPlayerCount: (uint8)(15),
	// })
	// (packet.UdpConnectionEstablished)(packet.UdpConnectionEstablished{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(0),
	// 		Type:   (packet.Type)(5),
	// 	}),
	// })
	// (packet.LoadLevel)(packet.LoadLevel{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(5),
	// 		Type:   (packet.Type)(20),
	// 	}),
	// 	LevelFilename: ([]uint8)([]uint8{
	// 		(uint8)(116),
	// 		(uint8)(101),
	// 		(uint8)(115),
	// 		(uint8)(116),
	// 		(uint8)(51),
	// 	}),
	// })
	// (string)("test3")
	// (packet.CurrentPlayersInfo)(packet.CurrentPlayersInfo{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(25),
	// 		Type:   (packet.Type)(21),
	// 	}),
	// 	Players: ([]packet.PlayerInfo)([]packet.PlayerInfo{
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(8),
	// 			Name: ([]uint8)([]uint8{
	// 				(uint8)(115),
	// 				(uint8)(104),
	// 				(uint8)(117),
	// 				(uint8)(114),
	// 				(uint8)(99),
	// 				(uint8)(111),
	// 				(uint8)(111),
	// 				(uint8)(76),
	// 			}),
	// 			Team:  (uint8)(2),
	// 			State: (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 		(packet.PlayerInfo)(packet.PlayerInfo{
	// 			NameLength: (uint8)(0),
	// 			Name:       ([]uint8)([]uint8{}),
	// 			Team:       (uint8)(0),
	// 			State:      (*packet.State)(nil),
	// 		}),
	// 	}),
	// })
	// (packet.EnterGamePermission)(packet.EnterGamePermission{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(0),
	// 		Type:   (packet.Type)(6),
	// 	}),
	// })
	// (packet.PlayerJoinedTeam)(packet.PlayerJoinedTeam{
	// 	TcpHeader: (packet.TcpHeader)(packet.TcpHeader{
	// 		Length: (uint16)(15),
	// 		Type:   (packet.Type)(28),
	// 	}),
	// 	PlayerId: (uint8)(0),
	// 	Team:     (uint8)(0),
	// 	State: (*packet.State)(&packet.State{
	// 		CommandSequenceNumber: (uint8)(123),
	// 		X: (float32)(1),
	// 		Y: (float32)(2),
	// 		Z: (float32)(3),
	// 	}),
	// })
	// done
}
