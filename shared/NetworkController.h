#pragma once
#ifndef __NetworkController_H__
#define __NetworkController_H__

class NetworkController
	: public PlayerController
{
public:
	NetworkController(CPlayer & oPlayer);
	~NetworkController();

	bool RequestInput(u_char cSequenceNumber);

	void ProcessCommand(CPacket & oPacket);

	u_char		cLastRecvedCommandSequenceNumber;
	bool		bFirstCommand;						// When true, indicates we are expecting the first command from a client (so far got nothing) and will be set to false when it arrives
	u_char		cCurrentCommandSeriesNumber;		// A number that changes on every respawn, team change, etc. and the server will ignore any Commands with mismatching series number

private:
	GLFWmutex		m_oMutex;		// DEBUG: What is this for? I forgot lol
};

#endif // __NetworkController_H__