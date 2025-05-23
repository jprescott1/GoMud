package hooks

import (
	"github.com/GoMudEngine/GoMud/internal/events"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/users"
)

// Checks whether their level is too high for a guide
func CheckGuide(e events.Event) events.ListenerReturn {

	evt, typeOk := e.(events.LevelUp)
	if !typeOk {
		mudlog.Error("Event", "Expected Type", "LevelUp", "Actual Type", e.Type())
		return events.Cancel
	}

	user := users.GetByUserId(evt.UserId)
	if user == nil {
		return events.Continue
	}

	if user.Character.Level >= 5 {
		for _, mobInstanceId := range user.Character.CharmedMobs {
			if mob := mobs.GetInstance(mobInstanceId); mob != nil {

				if mob.MobId == 38 {
					mob.Command(`say I see you have grown much stronger and more experienced. My assistance is now needed elsewhere. I wish you good luck!`)
					mob.Command(`emote clicks their heels together and disappears in a cloud of smoke.`, 10)
					mob.Command(`suicide vanish`, 10)
				}
			}
		}
	}

	return events.Continue
}
