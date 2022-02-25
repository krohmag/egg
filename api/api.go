package api

import (
	"context"
	"egg/datastore"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"

	"gorm.io/gorm"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	deviceId        = "IOS"
	clientVersion   = uint32(37)
	apiFirstContact = "https://www.auxbrain.com/ei/first_contact"
)

// GetBackupFromAPI queries the Egg, Inc. API for a user's backup info
func GetBackupFromAPI(eiUID string) (*FirstContact_Payload, error) {
	responseBody := new(FirstContact)
	authenticatedMsg := new(AuthenticatedMessage)
	payload := FirstContactRequestPayload{
		EiUserId:      eiUID,
		DeviceId:      deviceId,
		ClientVersion: clientVersion,
	}

	reqBin, err := proto.Marshal(proto.Message(&payload))
	if err != nil {
		return &FirstContact_Payload{}, err
	}

	reqDataEncoded := base64.StdEncoding.EncodeToString(reqBin)

	resp, err := http.PostForm(apiFirstContact, url.Values{"data": {reqDataEncoded}})
	defer func() {
		_ = resp.Body.Close()
	}()
	if err != nil {
		return &FirstContact_Payload{}, err
	}

	body, err := io.ReadAll(resp.Body)

	enc := base64.StdEncoding
	buf := make([]byte, enc.DecodedLen(len(body)))
	decoded, err := enc.Decode(buf, body)

	if err = proto.Unmarshal(buf[:decoded], authenticatedMsg); err != nil {
		return &FirstContact_Payload{}, err
	}

	if err = proto.Unmarshal(authenticatedMsg.Message, responseBody); err != nil {
		return &FirstContact_Payload{}, err
	}

	return responseBody.Data, nil
}

// AddUserToDatabase builds a datastore.User object and adds it to a datastore
func AddUserToDatabase(ctx context.Context, store datastore.Database, backup *FirstContact_Payload, discordName string) (datastore.User, error) {
	var soulFood int32
	var prophecyBonus int32
	for _, research := range backup.GetProgress().GetEpicResearches() {
		if research.Id == "soul_eggs" {
			soulFood = research.Level
		}
		if research.Id == "prophecy_bonus" {
			prophecyBonus = research.Level
		}
	}
	user := datastore.User{
		EggIncID:        backup.EiUserId,
		DiscordName:     discordName,
		GameAccountName: backup.UserName,
		SoulFood:        soulFood,
		ProphecyBonus:   prophecyBonus,
		SoulEggs:        backup.GetProgress().GetSoulEggs(),
		ProphecyEggs:    backup.GetProgress().GetProphecyEggs(),
	}

	tx, err := store.Transaction(ctx)
	if err != nil {
		return datastore.User{}, err
	}
	defer func() {
		if err == nil {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	record, err := tx.CreateOrUpdateUser(user)
	if err != nil {
		return datastore.User{}, err
	}

	return record, nil
}

// RemoveUserFromDatabase removes a user from the database provided the provided ID and discord username match up with the database record
func RemoveUserFromDatabase(ctx context.Context, store datastore.Database, eggID, discordName string) error {
	tx, err := store.Transaction(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	var check datastore.User
	switch record, checkErr := tx.GetUserByEggIncUserID(eggID); {
	case checkErr == nil:
		check = record
	case errors.Is(checkErr, gorm.ErrRecordNotFound):
		return errors.New("I don't have a record of you")
	default:
		return checkErr
	}
	if err != nil {
		return err
	}

	if check.DiscordName != discordName {
		return errors.New("Your Discord user is not associated with the ID you provided")
	}

	err = tx.DeleteUser(datastore.User{
		EggIncID:    eggID,
		DiscordName: discordName,
	})

	return err
}

func BuildSELeaderboard(ctx context.Context, store datastore.Database) (*discordgo.MessageEmbed, error) {
	tx, err := store.Transaction(ctx)
	if err != nil {
		return &discordgo.MessageEmbed{}, err
	}
	defer func() {
		if err == nil {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}()

	records, err := tx.GetUsers()
	if err != nil {
		return &discordgo.MessageEmbed{}, err
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].SoulEggs > records[j].SoulEggs
	})

	embedFields := make([]*discordgo.MessageEmbedField, 0)
	for i, record := range records {
		_, humanEB, se, mathErr := GetEBAndSE(record)
		if mathErr != nil {
			return &discordgo.MessageEmbed{}, mathErr
		}

		field := &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s %s", i+1, record.DiscordName, humanEB),
			Value:  fmt.Sprintf("%s soul eggs", se),
			Inline: false,
		}
		embedFields = append(embedFields, field)
	}

	embed := &discordgo.MessageEmbed{
		Type:      discordgo.EmbedTypeRich,
		Title:     "Soul Egg Leaderboard",
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x8700C3, // button purple
		// Color:     0x00ff00, // green
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprint("Updates every X minutes | Last updated"),
		},
		Fields: embedFields,
	}

	return embed, nil
}

// GetEBAndSE returns a calculated Earnings bonus as well as a count of Soul Eggs, both in a human readable format
func GetEBAndSE(user datastore.User) (float64, string, string, error) {
	rawEB, humanEB := calculateEB(user)
	return rawEB, humanEB, forPeople(user.SoulEggs), nil
}

func calculateEB(data datastore.User) (float64, string) {
	sePercent := (0.1 + (float64(data.SoulFood) * .01)) * 100
	pePercent := math.Pow(float64(1)+0.05+(float64(data.ProphecyBonus)*0.01), float64(data.ProphecyEggs)) * 100
	bonus := (sePercent * pePercent) / 100

	return bonus * data.SoulEggs, forPeople(bonus * data.SoulEggs)
}

func forPeople(bigAssNumber float64) string {
	units := []string{"", "k", "m", "b", "T", "q", "Q", "s", "S", "o", "N", "d"}
	k := float64(1000)
	magnitude := math.Floor(math.Log(bigAssNumber) / math.Log(k))
	return fmt.Sprintf("%.3f%s", bigAssNumber/(math.Pow(k, magnitude)), units[int(magnitude)])
}
