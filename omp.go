package omp

// #include <stdlib.h>
// #include "include/omp.h"
import "C"
import (
	"strings"
	"time"
	"unsafe"

	"github.com/kodeyeen/event"
)

type Animation struct {
	Lib, Name                  string
	Delta                      float32
	Loop, LockX, LockY, Freeze bool
	Duration                   time.Duration
}

type Vector4 struct {
	X float32
	Y float32
	Z float32
	W float32
}

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

type Vector2 struct {
	X float32
	Y float32
}

type Color uint

var events = event.NewDispatcher()
var commands = newCommandManager()

func On(_type event.Type, handler any) {
	events.On(_type, handler)
}

func Once(_type event.Type, handler any) {
	events.Once(_type, handler)
}

func Off(_type event.Type, handler any) {
	events.Off(_type, handler)
}

func Dispatch[T any](_type event.Type, data T) {
	event.Dispatch(events, _type, data)
}

func AddCommand(name string, handler CommandHandler) {
	commands.add(name, handler)
}

//export onGameModeInit
func onGameModeInit() {
	cLibPath := C.CString("./components/Gomponent.dll")
	defer C.free(unsafe.Pointer(cLibPath))

	C.init(cLibPath)

	event.Dispatch(events, EventTypeGameModeInit, &GameModeInitEvent{})
}

//export onGameModeExit
func onGameModeExit() {
	event.Dispatch(events, EventTypeGameModeExit, &GameModeExitEvent{})
}

// Actor events

//export onPlayerGiveDamageActor
func onPlayerGiveDamageActor(player, actor unsafe.Pointer, amount float32, weapon uint, part int) {
	event.Dispatch(events, EventTypePlayerGiveDamageActor, &PlayerGiveDamageActorEvent{
		Player: &Player{handle: player},
		Actor:  &Player{handle: actor},
		Amount: amount,
		Weapon: weapon,
		Part:   BodyPart(part),
	})
}

//export onActorStreamOut
func onActorStreamOut(actor, forPlayer unsafe.Pointer) {
	event.Dispatch(events, EventTypeActorStreamOut, &ActorStreamOutEvent{
		Actor:     &Player{handle: actor},
		ForPlayer: &Player{handle: forPlayer},
	})
}

//export onActorStreamIn
func onActorStreamIn(actor, forPlayer unsafe.Pointer) {
	event.Dispatch(events, EventTypeActorStreamIn, &ActorStreamInEvent{
		Actor:     &Player{handle: actor},
		ForPlayer: &Player{handle: forPlayer},
	})
}

// Checkpoint events

//export onPlayerEnterCheckpoint
func onPlayerEnterCheckpoint(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerEnterCheckpoint, &PlayerEnterCheckpointEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerLeaveCheckpoint
func onPlayerLeaveCheckpoint(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerLeaveCheckpoint, &PlayerLeaveCheckpointEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerEnterRaceCheckpoint
func onPlayerEnterRaceCheckpoint(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerEnterRaceCheckpoint, &PlayerEnterRaceCheckpointEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerLeaveRaceCheckpoint
func onPlayerLeaveRaceCheckpoint(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerLeaveRaceCheckpoint, &PlayerLeaveRaceCheckpointEvent{
		Player: &Player{handle: player},
	})
}

// Class events

//export onPlayerRequestClass
func onPlayerRequestClass(player, class unsafe.Pointer) bool {
	return event.Dispatch(events, EventTypePlayerRequestClass, &PlayerRequestClassEvent{
		Player: &Player{handle: player},
		Class:  &Class{handle: class},
	})
}

// Console events. TODO

//export onConsoleText
func onConsoleText(command C.String, parameters C.String) bool {
	return event.Dispatch(events, EventTypeConsoleText, &ConsoleTextEvent{
		Command:    C.GoStringN(command.buf, C.int(command.length)),
		Parameters: C.GoStringN(parameters.buf, C.int(parameters.length)),
	})
}

//export onRconLoginAttempt
func onRconLoginAttempt(player unsafe.Pointer, password C.String, success bool) {
	event.Dispatch(events, EventTypeRconLoginAttempt, &RconLoginAttemptEvent{
		Player:   &Player{handle: player},
		Password: C.GoStringN(password.buf, C.int(password.length)),
		Success:  success,
	})
}

// Custom model events

//export onPlayerFinishedDownloading
func onPlayerFinishedDownloading(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerFinishedDownloading, &PlayerFinishedDownloadingEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerRequestDownload
func onPlayerRequestDownload(player unsafe.Pointer, _type uint8, checksum uint32) bool {
	return event.Dispatch(events, EventTypePlayerRequestDownload, &PlayerRequestDownloadEvent{
		Player:   &Player{handle: player},
		Type:     int(_type),
		Checksum: int(checksum),
	})
}

// Dialog events

//export onDialogResponse
func onDialogResponse(player unsafe.Pointer, dialogID, response, listItem int, inputText C.String) {
	eventPlayer := &Player{handle: player}

	dialog := activeDialogs[eventPlayer.ID()]

	switch dialog := dialog.(type) {
	case *MessageDialog:
		event.Dispatch(dialog.Dispatcher, EventTypeDialogResponse, &MessageDialogResponseEvent{
			Player:   eventPlayer,
			Response: DialogResponse(response),
		})

		event.Dispatch(dialog.Dispatcher, EventTypeDialogHide, &DialogHideEvent{
			Player: eventPlayer,
		})
	case *InputDialog:
		event.Dispatch(dialog.Dispatcher, EventTypeDialogResponse, &InputDialogResponseEvent{
			Player:    eventPlayer,
			Response:  DialogResponse(response),
			InputText: C.GoStringN(inputText.buf, C.int(inputText.length)),
		})

		event.Dispatch(dialog.Dispatcher, EventTypeDialogHide, &DialogHideEvent{
			Player: eventPlayer,
		})
	case *ListDialog:
		event.Dispatch(dialog.Dispatcher, EventTypeDialogResponse, &ListDialogResponseEvent{
			Player:     eventPlayer,
			Response:   DialogResponse(response),
			ItemNumber: listItem,
			Item:       C.GoStringN(inputText.buf, C.int(inputText.length)),
		})

		event.Dispatch(dialog.Dispatcher, EventTypeDialogHide, &DialogHideEvent{
			Player: eventPlayer,
		})
	case *TabListDialog:
		event.Dispatch(dialog.Dispatcher, EventTypeDialogResponse, &TabListDialogResponseEvent{
			Player:     eventPlayer,
			Response:   DialogResponse(response),
			ItemNumber: listItem,
			Item:       dialog.items[listItem],
		})

		event.Dispatch(dialog.Dispatcher, EventTypeDialogHide, &DialogHideEvent{
			Player: eventPlayer,
		})
	}

	delete(activeDialogs, eventPlayer.ID())
}

// GangZone events

//export onPlayerEnterGangZone
func onPlayerEnterGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerEnterTurf, &PlayerEnterTurfEvent{
		Player: &Player{handle: player},
		Turf:   &Turf{handle: gangzone},
	})
}

//export onPlayerEnterPlayerGangZone
func onPlayerEnterPlayerGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerEnterPlayerTurf, &PlayerEnterPlayerTurfEvent{
		Player: &Player{handle: player},
		Turf:   &PlayerTurf{handle: gangzone},
	})
}

//export onPlayerLeaveGangZone
func onPlayerLeaveGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerLeaveTurf, &PlayerLeaveTurfEvent{
		Player: &Player{handle: player},
		Turf:   &Turf{handle: gangzone},
	})
}

//export onPlayerLeavePlayerGangZone
func onPlayerLeavePlayerGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerLeavePlayerTurf, &PlayerLeavePlayerTurfEvent{
		Player: &Player{handle: player},
		Turf:   &PlayerTurf{handle: gangzone},
	})
}

//export onPlayerClickGangZone
func onPlayerClickGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerClickTurf, &PlayerClickTurfEvent{
		Player: &Player{handle: player},
		Turf:   &Turf{handle: gangzone},
	})
}

//export onPlayerClickPlayerGangZone
func onPlayerClickPlayerGangZone(player, gangzone unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerClickPlayerTurf, &PlayerClickPlayerTurfEvent{
		Player: &Player{handle: player},
		Turf:   &PlayerTurf{handle: gangzone},
	})
}

// Menu events

//export onPlayerSelectedMenuRow
func onPlayerSelectedMenuRow(player unsafe.Pointer, menuRow uint8) {
	event.Dispatch(events, EventTypePlayerSelectedMenuRow, &PlayerSelectedMenuRowEvent{
		Player:  &Player{handle: player},
		MenuRow: menuRow,
	})
}

//export onPlayerExitedMenu
func onPlayerExitedMenu(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerExitedMenu, &PlayerExitedMenuEvent{
		Player: &Player{handle: player},
	})
}

// Object events

//export onObjectMoved
func onObjectMoved(object unsafe.Pointer) {
	event.Dispatch(events, EventTypeObjectMoved, &ObjectMovedEvent{
		Object: &Object{handle: object},
	})
}

//export onPlayerObjectMoved
func onPlayerObjectMoved(player, object unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerObjectMoved, &PlayerObjectMovedEvent{
		Player: &Player{handle: player},
		Object: &PlayerObject{handle: object},
	})
}

//export onObjectSelected
func onObjectSelected(player, object unsafe.Pointer, model int, pos C.Vector3) {
	event.Dispatch(events, EventTypeObjectSelected, &ObjectSelectedEvent{
		Player: &Player{handle: player},
		Object: &Object{handle: object},
		Model:  model,
		Position: Vector3{
			X: float32(pos.x),
			Y: float32(pos.y),
			Z: float32(pos.z),
		},
	})
}

//export onPlayerObjectSelected
func onPlayerObjectSelected(player, object unsafe.Pointer, model int, pos C.Vector3) {
	event.Dispatch(events, EventTypePlayerObjectSelected, &PlayerObjectSelectedEvent{
		Player: &Player{handle: player},
		Object: &PlayerObject{handle: object},
		Model:  model,
		Position: Vector3{
			X: float32(pos.x),
			Y: float32(pos.y),
			Z: float32(pos.z),
		},
	})
}

//export onObjectEdited
func onObjectEdited(player, object unsafe.Pointer, response int, offset, rot C.Vector3) {
	event.Dispatch(events, EventTypeObjectEdited, &ObjectEditedEvent{
		Player:   &Player{handle: player},
		Object:   &Object{handle: object},
		Response: ObjectEditResponse(response),
		Offset: Vector3{
			X: float32(offset.x),
			Y: float32(offset.y),
			Z: float32(offset.z),
		},
		Rotation: Vector3{
			X: float32(rot.x),
			Y: float32(rot.y),
			Z: float32(rot.z),
		},
	})
}

//export onPlayerObjectEdited
func onPlayerObjectEdited(player, object unsafe.Pointer, response int, offset, rot C.Vector3) {
	event.Dispatch(events, EventTypePlayerObjectEdited, &PlayerObjectEditedEvent{
		Player:   &Player{handle: player},
		Object:   &PlayerObject{handle: object},
		Response: ObjectEditResponse(response),
		Offset: Vector3{
			X: float32(offset.x),
			Y: float32(offset.y),
			Z: float32(offset.z),
		},
		Rotation: Vector3{
			X: float32(rot.x),
			Y: float32(rot.y),
			Z: float32(rot.z),
		},
	})
}

//export onPlayerAttachedObjectEdited
func onPlayerAttachedObjectEdited(player unsafe.Pointer, index int, saved bool, data C.PlayerAttachedObject) {
	event.Dispatch(events, EventTypePlayerAttachmentEdited, &PlayerAttachmentEdited{
		Player: &Player{handle: player},
		Index:  index,
		Saved:  saved,
		Attachment: PlayerAttachment{
			ModelID: int(data.model),
			Bone:    PlayerBone(data.bone),
			Offset: Vector3{
				X: float32(data.offset.x),
				Y: float32(data.offset.y),
				Z: float32(data.offset.z),
			},
			Rot: Vector3{
				X: float32(data.rotation.x),
				Y: float32(data.rotation.y),
				Z: float32(data.rotation.z),
			},
			Scale: Vector3{
				X: float32(data.scale.x),
				Y: float32(data.scale.y),
				Z: float32(data.scale.z),
			},
			Color1: Color(data.colour1),
			Color2: Color(data.colour2),
		},
	})
}

// Pickup events

//export onPlayerPickUpPickup
func onPlayerPickUpPickup(player, pickup unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerPickUpPickup, &PlayerPickUpPickupEvent{
		Player: &Player{handle: player},
		Pickup: &Pickup{handle: pickup},
	})
}

//export onPlayerPickUpPlayerPickup
func onPlayerPickUpPlayerPickup(player, pickup unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerPickUpPlayerPickup, &PlayerPickUpPlayerPickupEvent{
		Player: &Player{handle: player},
		Pickup: &PlayerPickup{handle: pickup},
	})
}

// Player spawn events

//export onPlayerRequestSpawn
func onPlayerRequestSpawn(player unsafe.Pointer) bool {
	return event.Dispatch(events, EventTypePlayerRequestSpawn, &PlayerRequestSpawnEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerSpawn
func onPlayerSpawn(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerSpawn, &PlayerSpawnEvent{
		Player: &Player{handle: player},
	})
}

// Player connect events

//export onIncomingConnection
func onIncomingConnection(player unsafe.Pointer, ipAddress C.String, port C.ushort) {
	event.Dispatch(events, EventTypeIncomingConnection, &IncomingConnectionEvent{
		Player:    &Player{handle: player},
		IPAddress: C.GoStringN(ipAddress.buf, C.int(ipAddress.length)),
		Port:      int(port),
	})
}

//export onPlayerConnect
func onPlayerConnect(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerConnect, &PlayerConnectEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerDisconnect
func onPlayerDisconnect(player unsafe.Pointer, reason int) {
	eventPlayer := &Player{handle: player}

	event.Dispatch(events, EventTypePlayerDisconnect, &PlayerDisconnectEvent{
		Player: eventPlayer,
		Reason: DisconnectReason(reason),
	})

	delete(activeDialogs, eventPlayer.ID())
}

//export onPlayerClientInit
func onPlayerClientInit(player unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerClientInit, &PlayerClientInitEvent{
		Player: &Player{handle: player},
	})
}

// Player stream events

//export onPlayerStreamIn
func onPlayerStreamIn(player, forPlayer unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerStreamIn, &PlayerStreamInEvent{
		Player:    &Player{handle: player},
		ForPlayer: &Player{handle: forPlayer},
	})
}

//export onPlayerStreamOut
func onPlayerStreamOut(player, forPlayer unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerStreamOut, &PlayerStreamOutEvent{
		Player:    &Player{handle: player},
		ForPlayer: &Player{handle: forPlayer},
	})
}

// Player text events

//export onPlayerText
func onPlayerText(player unsafe.Pointer, message *C.char) {
	event.Dispatch(events, EventTypePlayerText, &PlayerTextEvent{
		Player:  &Player{handle: player},
		Message: C.GoString(message),
	})
}

//export onPlayerCommandText
func onPlayerCommandText(player unsafe.Pointer, message C.String) bool {
	rawCmd := C.GoStringN(message.buf, C.int(message.length))

	tmp := strings.Fields(rawCmd)
	name := strings.TrimPrefix(tmp[0], "/")
	args := tmp[1:]

	exists := commands.has(name)
	if !exists {
		return false
	}

	commands.run(name, &Command{
		Sender:   &Player{handle: player},
		Name:     name,
		Args:     args,
		RawValue: rawCmd,
	})

	return true
}

// Player shot events

//export onPlayerShotMissed
func onPlayerShotMissed(player unsafe.Pointer, bulletData C.PlayerBulletData) bool {
	return event.Dispatch(events, EventTypePlayerShotMissed, &PlayerShotMissedEvent{
		Player: &Player{handle: player},
		Bullet: PlayerBullet{
			Origin: Vector3{
				X: float32(bulletData.origin.x),
				Y: float32(bulletData.origin.y),
				Z: float32(bulletData.origin.z),
			},
			HitPos: Vector3{
				X: float32(bulletData.hitPos.x),
				Y: float32(bulletData.hitPos.y),
				Z: float32(bulletData.hitPos.z),
			},
			Offset: Vector3{
				X: float32(bulletData.offset.x),
				Y: float32(bulletData.offset.y),
				Z: float32(bulletData.offset.z),
			},
			Weapon: Weapon(bulletData.weapon),
		},
	})
}

//export onPlayerShotPlayer
func onPlayerShotPlayer(player, target unsafe.Pointer, bulletData C.PlayerBulletData) bool {
	return event.Dispatch(events, EventTypePlayerShotPlayer, &PlayerShotPlayerEvent{
		Player: &Player{handle: player},
		Target: &Player{handle: target},
		Bullet: PlayerBullet{
			Origin: Vector3{
				X: float32(bulletData.origin.x),
				Y: float32(bulletData.origin.y),
				Z: float32(bulletData.origin.z),
			},
			HitPos: Vector3{
				X: float32(bulletData.hitPos.x),
				Y: float32(bulletData.hitPos.y),
				Z: float32(bulletData.hitPos.z),
			},
			Offset: Vector3{
				X: float32(bulletData.offset.x),
				Y: float32(bulletData.offset.y),
				Z: float32(bulletData.offset.z),
			},
			Weapon: Weapon(bulletData.weapon),
		},
	})
}

//export onPlayerShotVehicle
func onPlayerShotVehicle(player, target unsafe.Pointer, bulletData C.PlayerBulletData) bool {
	return event.Dispatch(events, EventTypePlayerShotVehicle, &PlayerShotVehicleEvent{
		Player: &Player{handle: player},
		Target: &Vehicle{handle: target},
		Bullet: PlayerBullet{
			Origin: Vector3{
				X: float32(bulletData.origin.x),
				Y: float32(bulletData.origin.y),
				Z: float32(bulletData.origin.z),
			},
			HitPos: Vector3{
				X: float32(bulletData.hitPos.x),
				Y: float32(bulletData.hitPos.y),
				Z: float32(bulletData.hitPos.z),
			},
			Offset: Vector3{
				X: float32(bulletData.offset.x),
				Y: float32(bulletData.offset.y),
				Z: float32(bulletData.offset.z),
			},
			Weapon: Weapon(bulletData.weapon),
		},
	})
}

//export onPlayerShotObject
func onPlayerShotObject(player, target unsafe.Pointer, bulletData C.PlayerBulletData) bool {
	return event.Dispatch(events, EventTypePlayerShotObject, &PlayerShotObjectEvent{
		Player: &Player{handle: player},
		Target: &Object{handle: target},
		Bullet: PlayerBullet{
			Origin: Vector3{
				X: float32(bulletData.origin.x),
				Y: float32(bulletData.origin.y),
				Z: float32(bulletData.origin.z),
			},
			HitPos: Vector3{
				X: float32(bulletData.hitPos.x),
				Y: float32(bulletData.hitPos.y),
				Z: float32(bulletData.hitPos.z),
			},
			Offset: Vector3{
				X: float32(bulletData.offset.x),
				Y: float32(bulletData.offset.y),
				Z: float32(bulletData.offset.z),
			},
			Weapon: Weapon(bulletData.weapon),
		},
	})
}

//export onPlayerShotPlayerObject
func onPlayerShotPlayerObject(player, target unsafe.Pointer, bulletData C.PlayerBulletData) bool {
	return event.Dispatch(events, EventTypePlayerShotPlayerObject, &PlayerShotPlayerObjectEvent{
		Player: &Player{handle: player},
		Target: &PlayerObject{handle: target},
		Bullet: PlayerBullet{
			Origin: Vector3{
				X: float32(bulletData.origin.x),
				Y: float32(bulletData.origin.y),
				Z: float32(bulletData.origin.z),
			},
			HitPos: Vector3{
				X: float32(bulletData.hitPos.x),
				Y: float32(bulletData.hitPos.y),
				Z: float32(bulletData.hitPos.z),
			},
			Offset: Vector3{
				X: float32(bulletData.offset.x),
				Y: float32(bulletData.offset.y),
				Z: float32(bulletData.offset.z),
			},
			Weapon: Weapon(bulletData.weapon),
		},
	})
}

// Player change events

//export onPlayerScoreChange
func onPlayerScoreChange(player unsafe.Pointer, score int) {
	event.Dispatch(events, EventTypePlayerScoreChange, &PlayerScoreChangeEvent{
		Player: &Player{handle: player},
		Score:  score,
	})
}

//export onPlayerNameChange
func onPlayerNameChange(player unsafe.Pointer, oldName C.String) {
	event.Dispatch(events, EventTypePlayerNameChange, &PlayerNameChangeEvent{
		Player:  &Player{handle: player},
		OldName: C.GoStringN(oldName.buf, C.int(oldName.length)),
	})
}

//export onPlayerInteriorChange
func onPlayerInteriorChange(player unsafe.Pointer, newInterior, oldInterior uint) {
	event.Dispatch(events, EventTypePlayerInteriorChange, &PlayerInteriorChangeEvent{
		Player:      &Player{handle: player},
		NewInterior: newInterior,
		OldInterior: oldInterior,
	})
}

//export onPlayerStateChange
func onPlayerStateChange(player unsafe.Pointer, newState, oldState int) {
	event.Dispatch(events, EventTypePlayerStateChange, &PlayerStateChangeEvent{
		Player:   &Player{handle: player},
		NewState: PlayerState(newState),
		OldState: PlayerState(oldState),
	})
}

//export onPlayerKeyStateChange
func onPlayerKeyStateChange(player unsafe.Pointer, newKeys, oldKeys uint) {
	event.Dispatch(events, EventTypePlayerKeyStateChange, &PlayerKeyStateChangeEvent{
		Player:  &Player{handle: player},
		NewKeys: newKeys,
		OldKeys: oldKeys,
	})
}

// Player damage events

//export onPlayerDeath
func onPlayerDeath(player, killer unsafe.Pointer, reason int) {
	eventKiller := &Player{handle: killer}
	if killer == nil {
		eventKiller = nil
	}

	event.Dispatch(events, EventTypePlayerDeath, &PlayerDeathEvent{
		Player: &Player{handle: player},
		Killer: eventKiller,
		Reason: reason,
	})
}

//export onPlayerTakeDamage
func onPlayerTakeDamage(player, from unsafe.Pointer, amount float32, weapon uint, part int) {
	eventFrom := &Player{handle: from}
	if from == nil {
		eventFrom = nil
	}

	event.Dispatch(events, EventTypePlayerTakeDamage, &PlayerTakeDamageEvent{
		Player: &Player{handle: player},
		From:   eventFrom,
		Amount: amount,
		Weapon: Weapon(weapon),
		Part:   BodyPart(part),
	})
}

//export onPlayerGiveDamage
func onPlayerGiveDamage(player, to unsafe.Pointer, amount float32, weapon uint, part int) {
	event.Dispatch(events, EventTypePlayerGiveDamage, &PlayerGiveDamageEvent{
		Player: &Player{handle: player},
		To:     &Player{handle: to},
		Amount: amount,
		Weapon: Weapon(weapon),
		Part:   BodyPart(part),
	})
}

// Player click events

//export onPlayerClickMap
func onPlayerClickMap(player unsafe.Pointer, pos C.Vector3) {
	event.Dispatch(events, EventTypePlayerClickMap, &PlayerClickMapEvent{
		Player: &Player{handle: player},
		Position: Vector3{
			X: float32(pos.x),
			Y: float32(pos.y),
			Z: float32(pos.z),
		},
	})
}

//export onPlayerClickPlayer
func onPlayerClickPlayer(player, clicked unsafe.Pointer, source int) {
	event.Dispatch(events, EventTypePlayerClickPlayer, &PlayerClickPlayerEvent{
		Player:  &Player{handle: player},
		Clicked: &Player{handle: clicked},
		Source:  PlayerClickSource(source),
	})
}

// Player check events

//export onClientCheckResponse
func onClientCheckResponse(player unsafe.Pointer, actionType, address, results int) {
	event.Dispatch(events, EventTypeClientCheckResponse, &ClientCheckResponseEvent{
		Player:     &Player{handle: player},
		ActionType: actionType,
		Address:    address,
		Results:    results,
	})
}

// Player update event

//export onPlayerUpdate
func onPlayerUpdate(player unsafe.Pointer, now C.longlong) bool {
	return event.Dispatch(events, EventTypePlayerUpdate, &PlayerUpdateEvent{
		Player: &Player{handle: player},
		Now:    time.Unix(0, int64(now)*int64(time.Millisecond)),
	})
}

// Textdraw events

//export onPlayerClickTextDraw
func onPlayerClickTextDraw(player, textdraw unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerClickTextDraw, &PlayerClickTextDrawEvent{
		Player:   &Player{handle: player},
		Textdraw: &Textdraw{handle: textdraw},
	})
}

//export onPlayerClickPlayerTextDraw
func onPlayerClickPlayerTextDraw(player, textdraw unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerClickPlayerTextDraw, &PlayerClickPlayerTextDrawEvent{
		Player:   &Player{handle: player},
		Textdraw: &PlayerTextdraw{handle: textdraw},
	})
}

//export onPlayerCancelTextDrawSelection
func onPlayerCancelTextDrawSelection(player unsafe.Pointer) bool {
	return event.Dispatch(events, EventTypePlayerCancelTextDrawSelection, &PlayerCancelTextDrawSelectionEvent{
		Player: &Player{handle: player},
	})
}

//export onPlayerCancelPlayerTextDrawSelection
func onPlayerCancelPlayerTextDrawSelection(player unsafe.Pointer) bool {
	return event.Dispatch(events, EventTypePlayerCancelPlayerTextDrawSelection, &PlayerCancelPlayerTextDrawSelectionEvent{
		Player: &Player{handle: player},
	})
}

// Vehicle events

//export onVehicleStreamIn
func onVehicleStreamIn(vehicle, player unsafe.Pointer) {
	event.Dispatch(events, EventTypeVehicleStreamIn, &VehicleStreamInEvent{
		Vehicle:   &Vehicle{handle: vehicle},
		ForPlayer: &Player{handle: player},
	})
}

//export onVehicleStreamOut
func onVehicleStreamOut(vehicle, player unsafe.Pointer) {
	event.Dispatch(events, EventTypeVehicleStreamOut, &VehicleStreamOutEvent{
		Vehicle:   &Vehicle{handle: vehicle},
		ForPlayer: &Player{handle: player},
	})
}

//export onVehicleDeath
func onVehicleDeath(vehicle, killer unsafe.Pointer) {
	event.Dispatch(events, EventTypeVehicleDeath, &VehicleDeathEvent{
		Vehicle: &Vehicle{handle: vehicle},
		Killer:  &Player{handle: killer},
	})
}

//export onPlayerEnterVehicle
func onPlayerEnterVehicle(player, vehicle unsafe.Pointer, isPassenger int) {
	event.Dispatch(events, EventTypePlayerEnterVehicle, &PlayerEnterVehicleEvent{
		Player:      &Player{handle: player},
		Vehicle:     &Vehicle{handle: vehicle},
		IsPassenger: isPassenger != 0,
	})
}

//export onPlayerExitVehicle
func onPlayerExitVehicle(player, vehicle unsafe.Pointer) {
	event.Dispatch(events, EventTypePlayerExitVehicle, &PlayerExitVehicleEvent{
		Player:  &Player{handle: player},
		Vehicle: &Vehicle{handle: vehicle},
	})
}

//export onVehicleDamageStatusUpdate
func onVehicleDamageStatusUpdate(vehicle, player unsafe.Pointer) {
	event.Dispatch(events, EventTypeVehicleDamageStatusUpdate, &VehicleDamageStatusUpdateEvent{
		Vehicle: &Vehicle{handle: vehicle},
		Player:  &Player{handle: player},
	})
}

//export onVehiclePaintJob
func onVehiclePaintJob(player, vehicle unsafe.Pointer, paintJob int) {
	event.Dispatch(events, EventTypeVehiclePaintJob, &VehiclePaintJobEvent{
		Player:   &Player{handle: player},
		Vehicle:  &Vehicle{handle: vehicle},
		PaintJob: paintJob,
	})
}

//export onVehicleMod
func onVehicleMod(player, vehicle unsafe.Pointer, component int) {
	event.Dispatch(events, EventTypeVehicleMod, &VehicleModEvent{
		Player:    &Player{handle: player},
		Vehicle:   &Vehicle{handle: vehicle},
		Component: component,
	})
}

//export onVehicleRespray
func onVehicleRespray(player, vehicle unsafe.Pointer, color1, color2 int) {
	event.Dispatch(events, EventTypeVehicleRespray, &VehicleResprayEvent{
		Player:  &Player{handle: player},
		Vehicle: &Vehicle{handle: vehicle},
		Color:   VehicleColor{Primary: color1, Secondary: color2},
	})
}

//export onEnterExitModShop
func onEnterExitModShop(player unsafe.Pointer, enterexit bool, interiorID int) {
	event.Dispatch(events, EventTypeEnterExitModShop, &EnterExitModShopEvent{
		Player:     &Player{handle: player},
		EnterExit:  enterexit,
		InteriorID: interiorID,
	})
}

//export onVehicleSpawn
func onVehicleSpawn(vehicle unsafe.Pointer) {
	event.Dispatch(events, EventTypeVehicleSpawn, &VehicleSpawnEvent{
		Vehicle: &Vehicle{handle: vehicle},
	})
}

//export onUnoccupiedVehicleUpdate
func onUnoccupiedVehicleUpdate(vehicle, player unsafe.Pointer, updateData C.UnoccupiedVehicleUpdate) bool {
	return event.Dispatch(events, EventTypeUnoccupiedVehicleUpdate, &UnoccupiedVehicleUpdateEvent{
		Vehicle: &Vehicle{handle: vehicle},
		Player:  &Player{handle: player},
		Update: UnoccupiedVehicleUpdate{
			Seat: int(updateData.seat),
			Position: Vector3{
				X: float32(updateData.position.x),
				Y: float32(updateData.position.y),
				Z: float32(updateData.position.z),
			},
			Velocity: Vector3{
				X: float32(updateData.velocity.x),
				Y: float32(updateData.velocity.y),
				Z: float32(updateData.velocity.z),
			},
		},
	})
}

//export onTrailerUpdate
func onTrailerUpdate(player, vehicle unsafe.Pointer) bool {
	return event.Dispatch(events, EventTypeTrailerUpdate, &TrailerUpdateEvent{
		Player:  &Player{handle: player},
		Vehicle: &Vehicle{handle: vehicle},
	})
}

//export onVehicleSirenStateChange
func onVehicleSirenStateChange(player, vehicle unsafe.Pointer, sirenState uint8) bool {
	return event.Dispatch(events, EventTypeVehicleSirenStateChange, &VehicleSirenStateChangeEvent{
		Player:     &Player{handle: player},
		Vehicle:    &Vehicle{handle: vehicle},
		SirenState: int(sirenState),
	})
}
