package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

type CommandData struct {
	Name string

	HumanName   string
	Description string
}

type Command interface {
	CommandData() CommandData
}

func InitFor(c Command, s *state.State) {
	r := reflect.ValueOf(c).MethodByName("Invoke")
	inputKind := r.Type().In(1)
	ctx := RequestContext{s}
	ctxValue := reflect.ValueOf(ctx)
	_, mapping := optionsFromType(inputKind)

	s.AddHandler(func(m *gateway.InteractionCreateEvent) {
		commandInteraction, ok := m.Data.(*discord.CommandInteraction)
		if !ok {
			return
		}

		requestValue := reflect.New(inputKind).Elem()

		for _, opt := range commandInteraction.Options {
			field := requestValue.FieldByName(mapping[opt.Name])

			switch field.Type() {
			case rString:
				field.SetString(opt.String())
			case rInt:
				intv, _ := opt.IntValue()
				field.SetInt(intv)
			case rBool:
				boolv, _ := opt.BoolValue()
				field.SetBool(boolv)
			case rUserID, rChannelID, rRoleID:
				snowv, _ := opt.SnowflakeValue()
				field.SetInt(int64(snowv))
			default:
				panic("unsupported type " + field.Type().String())
			}
		}

		r.Call([]reflect.Value{ctxValue, requestValue})
	})
}

type ExampleRequest struct {
	Name discord.UserID `req:"optional" desc:"your mom's name"`
	Role discord.RoleID `req:"optional" desc:"uoooh"`
}

type RequestContext struct {
	Discord *state.State
}

var (
	rString    = reflect.TypeOf("")
	rInt       = reflect.TypeOf(int(0))
	rBool      = reflect.TypeOf(false)
	rUserID    = reflect.TypeOf(discord.UserID(0))
	rChannelID = reflect.TypeOf(discord.ChannelID(0))
	rRoleID    = reflect.TypeOf(discord.RoleID(0))
)

func optionsFromType(t reflect.Type) (opts discord.CommandOptions, commandValueNamesToStructFields map[string]string) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		desc := field.Tag.Get("desc")
		required := field.Tag.Get("req") != "optional"

		switch field.Type {
		case rString:
			opts = append(opts, discord.NewBooleanOption(name, desc, required))
		case rInt:
			opts = append(opts, discord.NewIntegerOption(name, desc, required))
		case rBool:
			opts = append(opts, discord.NewBooleanOption(name, desc, required))
		case rUserID:
			opts = append(opts, discord.NewUserOption(name, desc, required))
		case rChannelID:
			opts = append(opts, discord.NewChannelOption(name, desc, required))
		case rRoleID:
			opts = append(opts, discord.NewRoleOption(name, desc, required))
		default:
			panic("unsupported type " + field.Type.String())
		}
	}
	return opts, commandValueNamesToStructFields
}

func OptionsFromRequest(v interface{}) (opts discord.CommandOptions, commandValueNamesToStructFields map[string]string) {
	r := reflect.ValueOf(v)

	return optionsFromType(r.Type())
}

func main() {
	v, _ := OptionsFromRequest(ExampleRequest{})
	d, _ := json.Marshal(v)
	fmt.Printf("%s", d)
}
