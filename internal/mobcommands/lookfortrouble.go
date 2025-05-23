package mobcommands

import (
	"fmt"

	"github.com/GoMudEngine/GoMud/internal/buffs"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/parties"
	"github.com/GoMudEngine/GoMud/internal/races"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

func LookForTrouble(rest string, mob *mobs.Mob, room *rooms.Room) (bool, error) {

	// Already aggroed, skip.
	if mob.Character.Aggro != nil {
		return true, nil
	}

	// Make a list of all players this gorup is hostile to in this room.
	isCharmed := mob.Character.IsCharmed()

	allPotentialTargets := []int{}
	nonDownedUserTargets := []int{}
	possibleMobTargets := []int{}

	//mudlog.Info("lookgfortrouble", "mobname", fmt.Sprintf(`%s#%d`, mob.Character.Name, mob.InstanceId))

	if !isCharmed {

		allPlayerIds := room.GetPlayers(rooms.FindAll)

		for _, playerId := range allPlayerIds {

			user := users.GetByUserId(playerId)
			if user == nil {
				continue
			}

			raceInfo := races.GetRace(user.Character.RaceId)
			if raceInfo == nil {
				mudlog.Error("RaceError", "Not Found", user.Character.RaceId)
				continue
			}

			// Once a player is downed mobs stop considering them a target
			// They don't see players that are sneaking...
			ignoreUser := false

			if user.Character.Health < 1 {
				ignoreUser = true
			} else if user.Character.HasBuffFlag(buffs.Hidden) {
				ignoreUser = true
			}

			entries := 1
			if party := parties.Get(playerId); party != nil {
				entries += party.ChanceToBeTargetted(playerId)
			}

			if mob.Hostile { // Does it always attack players?

				allPotentialTargets = append(allPotentialTargets, playerId)

				if !ignoreUser {
					for i := 0; i < entries; i++ {
						nonDownedUserTargets = append(nonDownedUserTargets, playerId)
					}
				}
				continue
			}

			// Does this specific mob hate this player?
			if mob.HatesRace(raceInfo.Name) || mob.HatesAlignment(user.Character.Alignment) {

				allPotentialTargets = append(allPotentialTargets, playerId)

				if !ignoreUser {
					for i := 0; i < entries; i++ {
						nonDownedUserTargets = append(nonDownedUserTargets, playerId)
					}
				}
				continue
			}

			// Hostility default to 5 minutes
			for _, groupName := range mob.Groups {
				// Does this group hate this player?
				if mobs.IsHostile(groupName, playerId) {

					allPotentialTargets = append(allPotentialTargets, playerId)

					if !ignoreUser {
						for i := 0; i < entries; i++ {
							nonDownedUserTargets = append(nonDownedUserTargets, playerId)
						}
					}
					break
				}
			}
		}

		// Still nothing, look for trouble in mobs they hate
		for _, considerMobInstanceId := range room.GetMobs() {
			if mob.InstanceId == considerMobInstanceId {
				continue
			}

			considerMob := mobs.GetInstance(considerMobInstanceId)
			if considerMob == nil {
				continue
			}

			raceInfo := races.GetRace(mob.Character.RaceId)

			if mob.HatesMob(considerMob) || mob.HatesRace(raceInfo.Name) {
				possibleMobTargets = append(possibleMobTargets, considerMobInstanceId)
				continue
			}

			if considerMob.Character.IsCharmed() {
				for _, uid := range allPotentialTargets { // Only consider charmed mobs if they are charmed by a player this mob wants to fight
					if considerMob.Character.IsCharmed(uid) {
						possibleMobTargets = append(possibleMobTargets, considerMobInstanceId)
						break
					}
				}
			}

		}
	}

	targetUserId := 0
	targetMobInstanceId := 0

	userCt := len(nonDownedUserTargets)
	mobCt := len(possibleMobTargets)

	if userCt > 0 || mobCt > 0 {
		randRoll := util.Rand(userCt + mobCt)
		if randRoll < userCt {
			targetUserId = nonDownedUserTargets[randRoll]
		} else {
			targetMobInstanceId = possibleMobTargets[randRoll-userCt]
		}
	}

	if targetUserId > 0 {
		mob.Command(fmt.Sprintf("attack @%d", targetUserId)) // @ denotes a specific player id
	} else if targetMobInstanceId > 0 {
		mob.Command(fmt.Sprintf("attack #%d", targetMobInstanceId)) // # denotes a specific mob id
	} else {

		if mob.Despawns() {
			if mob.BoredomCounter < 255 {
				mob.BoredomCounter++
			}
		}
	}

	return true, nil
}
