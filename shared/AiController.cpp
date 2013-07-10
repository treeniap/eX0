// TODO: Properly fix this, by making this file independent of globals.h
#ifdef EX0_CLIENT
#	include "../eX0mp/src/globals.h"
#else
#	include "../eX0ds/src/globals.h"
#endif // EX0_CLIENT

AiController::AiController(CPlayer & oPlayer)
	: LocalController(oPlayer)
{
}

AiController::~AiController()
{
}

void AiController::ProvideRealtimeInput(double dTimePassed)
{
	m_oPlayer.Rotate(0.4f * static_cast<float>(dTimePassed));
}

//bool AiController::RequestCommand(u_char/* cSequenceNumber*/)
void AiController::ProvideNextCommand()
{
	SequencedCommand_t oSequencedCommand;

	// Run around in circles
	oSequencedCommand.oCommand.cMoveDirection = 0;
	oSequencedCommand.oCommand.cStealth = 1;
	oSequencedCommand.oCommand.fZ = m_oPlayer.GetZ();
	// eDRODx AI
	/*m_oPlayer.SetStealth(false);
	m_oPlayer.MoveDirection(pLocalPlayer->GetMoveDirection() != -1 && !pLocalPlayer->IsDead() ? 0 : -1);
	if (!pLocalPlayer->IsDead()) {
		m_oPlayer.UpdateInterpolatedPos();
		m_oPlayer.Rotate(MathCoordToRad(static_cast<int>(m_oPlayer.GetIntX()),
										static_cast<int>(m_oPlayer.GetIntY()),
										static_cast<int>(pLocalPlayer->GetIntX()),
										static_cast<int>(pLocalPlayer->GetIntY())) + Math::HALF_PI - m_oPlayer.GetZ());
	}*/

	oSequencedCommand.cSequenceNumber = g_cCurrentCommandSequenceNumber;

	eX0_assert(m_oPlayer.m_oInputCmdsTEST.push(oSequencedCommand), "m_oInputCmdsTEST.push(oCommand) failed, lost a command!!\n");
}
