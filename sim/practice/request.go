//go:build practice

package practice

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/wowsims/classic/sim/core"
	"github.com/wowsims/classic/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

// ParseRequest reads what the page got from the class sim: a RaidSimRequest as JSON, an
// IndividualSimSettings as JSON (the class sim's JSON export), or the class sim's export link,
// whose hash holds the settings as zlib compressed proto in base64.
func ParseRequest(data []byte) (*proto.RaidSimRequest, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return nil, errors.New("empty settings")
	}

	if !strings.HasPrefix(text, "{") {
		settings, err := decodeLink(text)
		if err != nil {
			return nil, err
		}
		return requestFromSettings(settings), nil
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &probe); err != nil {
		return nil, err
	}
	unmarshal := protojson.UnmarshalOptions{DiscardUnknown: true}
	if _, isRequest := probe["raid"]; isRequest {
		request := &proto.RaidSimRequest{}
		if err := unmarshal.Unmarshal([]byte(text), request); err != nil {
			return nil, err
		}
		return request, nil
	}
	settings := &proto.IndividualSimSettings{}
	if err := unmarshal.Unmarshal([]byte(text), settings); err != nil {
		return nil, err
	}
	return requestFromSettings(settings), nil
}

// decodeLink reads the settings out of a class sim link (".../enhancement_shaman/#eJz...") or
// out of the hash alone.
func decodeLink(link string) (*proto.IndividualSimSettings, error) {
	hash := link
	if i := strings.Index(link, "#"); i >= 0 {
		hash = link[i+1:]
	}
	if unescaped, err := url.QueryUnescape(hash); err == nil {
		hash = unescaped
	}
	raw, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		if raw, err = base64.RawURLEncoding.DecodeString(hash); err != nil {
			return nil, fmt.Errorf("not a class sim link: %w", err)
		}
	}
	reader, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("not a class sim link: %w", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	settings := &proto.IndividualSimSettings{}
	if err := goproto.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func requestFromSettings(settings *proto.IndividualSimSettings) *proto.RaidSimRequest {
	request := &proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs),
		Encounter:  settings.Encounter,
		SimOptions: &proto.SimOptions{},
	}
	request.Raid.Tanks = settings.Tanks
	if settings.Settings != nil {
		request.SimOptions.Ruleset = settings.Settings.Ruleset
	}
	return request
}

// Player is your unit's proto in the request: the first player of the first party.
func Player(request *proto.RaidSimRequest) *proto.Player {
	for _, party := range request.GetRaid().GetParties() {
		for _, player := range party.GetPlayers() {
			if player != nil && player.Class != proto.Class_ClassUnknown {
				return player
			}
		}
	}
	return nil
}

// Normalize makes the changes the room needs to a request, on a copy. It never changes how
// the sim plays a unit, only what would act for us:
//
//   - Your rotation is an empty APL, with no priority list and no prepull actions. A nil
//     rotation panics, and the imported prepull actions would cast for you, because the APL
//     registers them when the sim is built.
//   - The fight has no health end and no length variation, so its length is exactly the
//     mode's.
//   - PvP melee downtime is 0, because it would move you in and out of range on its own. The
//     world decides range.
//
// Raid hits, the tank and the party cooldowns stay as in the request.
func Normalize(request *proto.RaidSimRequest) (*proto.RaidSimRequest, error) {
	request = goproto.Clone(request).(*proto.RaidSimRequest)
	if request.Encounter == nil || len(request.Encounter.Targets) == 0 {
		return nil, errors.New("the settings have no encounter target")
	}
	if request.Encounter.Leveling != nil {
		return nil, errors.New("a leveling encounter can't be practiced")
	}
	player := Player(request)
	if player == nil {
		return nil, errors.New("the settings have no player")
	}
	player.Rotation = &proto.APLRotation{Type: proto.APLRotation_TypeAPL}

	request.Encounter.UseHealth = false
	request.Encounter.DurationVariation = 0
	request.Encounter.PvpMeleeDowntime = 0
	if request.SimOptions == nil {
		request.SimOptions = &proto.SimOptions{}
	}
	request.SimOptions.Iterations = 1
	request.SimOptions.Debug = false
	request.SimOptions.DebugFirstIteration = false
	request.SimOptions.Interactive = false
	request.SimOptions.ReplaySeeds = nil
	return request, nil
}
