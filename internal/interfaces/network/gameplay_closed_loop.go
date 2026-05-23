package network

import (
	"context"

	equipmentv1 "github.com/slimeyquest/proto/gen/go/equipment"
	gatewayv1 "github.com/slimeyquest/proto/gen/go/gateway"
	playerv1 "github.com/slimeyquest/proto/gen/go/player"
)

func (g *Gameplay) handleCreateRole(c *Conn, req *playerv1.CreateRoleReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.CreateRole(context.Background(), c.PlayerID(), req.GetDisplayName())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_CreateRole{CreateRole: res}})
}

func (g *Gameplay) handleChestOpen(c *Conn, req *equipmentv1.ChestOpenReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.OpenChest(context.Background(), c.PlayerID(), req.GetCount())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_ChestOpen{ChestOpen: res}})
}

func (g *Gameplay) handleEquipItem(c *Conn, req *equipmentv1.EquipItemReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.EquipItem(context.Background(), c.PlayerID(), req.GetEquipmentUid(), req.GetSlot())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_EquipItem{EquipItem: res}})
}

func (g *Gameplay) handleDecomposeEquipment(c *Conn, req *equipmentv1.DecomposeEquipmentReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.DecomposeEquipment(context.Background(), c.PlayerID(), req.GetEquipmentUid())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_DecomposeEquipment{DecomposeEquipment: res}})
}

func (g *Gameplay) handleUpgradeChest(c *Conn, req *equipmentv1.UpgradeChestReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.UpgradeChest(context.Background(), c.PlayerID(), req.GetTargetLevel())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_UpgradeChest{UpgradeChest: res}})
}

func (g *Gameplay) handleDrawSkill(c *Conn, req *playerv1.DrawSkillReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.DrawSkill(context.Background(), c.PlayerID(), req.GetDrawCount())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_DrawSkill{DrawSkill: res}})
}

func (g *Gameplay) handleDrawCompanion(c *Conn, req *playerv1.DrawCompanionReq) error {
	if !g.session.Validate(c.PlayerID(), c.Token()) {
		return errInvalidSession
	}
	res, err := g.loop.DrawCompanion(context.Background(), c.PlayerID(), req.GetDrawCount())
	if err != nil {
		return err
	}
	return c.sendServerMessage(&gatewayv1.ServerMessage{Payload: &gatewayv1.ServerMessage_DrawCompanion{DrawCompanion: res}})
}
