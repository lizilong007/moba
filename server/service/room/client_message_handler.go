package room

import (
	"fmt"
	"github.com/golang/glog"
	"math/rand"

	"github.com/yangsf5/moba/server/hero"
	"github.com/yangsf5/moba/server/proto"
	"github.com/yangsf5/moba/server/user"
)

func (r *Room) HandleClientMessage(session int, msgType string, msgData interface{}) {
	u, ok := r.sessions[session]
	if !ok {
		glog.Errorf("RoomService-%s HallClientMessage session not in room, sessionId=%d", r.serviceName, session)
		return
	}

	switch msgType {
	case "chooseHero":
		if r.battleStatus != "waiting" {
			break
		}
		heroID := int(msgData.(float64))
		// TODO check heroID

		u.SetHeroID(heroID)

		retMsg := &proto.RCChooseHeroRet{1}
		u.Send([]byte(proto.Encode(r.serviceName, retMsg)))

		// 每当有人选好英雄就计算是否够人开打
		r.UpdateBattleStatus()
		r.NotifyRCPlayerInfos()
	case "shoot":
		if u.IsDead() {
			break
		}
		targetName := msgData.(string)
		targetUser := r.group.GetPeer(targetName)
		if targetUser == nil {
			break
		}
		if r.battleStatus != "firing" {
			break
		}
		mobaUser := targetUser.(User)
		if mobaUser.IsDead() {
			break
		}

		hp := mobaUser.GetHP() - (1 + rand.Intn(10))
		if hp <= 0 {
			hp = 0
		}
		mobaUser.SetHP(hp)
		shootMsg := &proto.RCShoot{u.Name(), targetName, hp}
		r.Broadcast(proto.Encode(r.serviceName, shootMsg))
	case "move":
		if u.IsDead() {
			break
		}
		position := msgData.(string)
		var x, y int
		fmt.Sscanf(position, "%d,%d", &x, &y)
		// TODO check x, y
		u.SetPosition(x, y)
		moveMsg := &proto.RCMove{u.Name(), x, y}
		r.Broadcast(proto.Encode(r.serviceName, moveMsg))
	case "skill":
		if r.battleStatus != "firing" {
			break
		}
		skillID := int(msgData.(float64))

		// TODO 根据离玩家为之最近或者伤害最低者来取目标玩家
		targetName := ""
		// 这里的目标玩家为空没关系，某些技能是非单体的，不需要单个目标
		targetPeer := r.group.GetPeer(targetName)
		var targetUser *user.User
		if targetPeer == nil {
			targetUser = nil
		} else {
			targetUser = targetPeer.(*user.User)
		}

		heroID := u.GetHeroID()
		err := hero.DoSkill(r.serviceName, u.(*user.User), targetUser, r.GetGroup(), heroID, skillID)
		if err != nil {
			glog.Errorf("RoomService-%s HallClientMessage do skill error, sessionId=%d error=[%v]", r.serviceName, session, err)
		}
	}
}
