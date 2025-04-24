package getdictentry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/model/oxf"
	"aidicti.top/scn/internal/model"
	"aidicti.top/scn/internal/service/scenario"
)

type scnr struct {
	usr scenario.DialogUSR
	stt scenario.DialogSTT
	ai  scenario.DialogGPT
	oxf scenario.DialogOXF
	uis scenario.DialogUIS

	name string
}

func New(
	stt scenario.DialogSTT,
	usr scenario.DialogUSR,
	ai scenario.DialogGPT,
	oxf scenario.DialogOXF,
	uis scenario.DialogUIS) *scnr {

	return &scnr{usr: usr, stt: stt, ai: ai, oxf: oxf, uis: uis, name: "getdictentry"}
}

func (s scnr) Name() string {
	return s.name
}

func (s scnr) Run(ctx context.Context) {

	// select {
	// case s.usr.Reqs() <- model.Message{Message: "Please write a word"}:
	// 	logging.Debug("send req to usr", "st", "ok", "type", s.name)
	// case <-ctx.Done():
	// 	logging.Debug("send req to usr", "st", "fail", "rsn", "ctx done", "type", s.name)
	// 	return
	// }

	select {
	case answ := <-s.usr.Resps():
		logging.Debug("get req from usr", "st", "ok", "respId", answ.Id, "respMsg", answ.Message, "type", s.name)

		for answ.Message == "" {
			logging.Debug("resend req to usr", "st", "ok", "rsn", "empty data", "type", s.name)

			s.usr.Reqs() <- model.Message{Message: "Please write something"}

			select {
			case answ = <-s.usr.Resps():
				logging.Debug("get req from usr", "st", "ok", "respId", answ.Id, "respMsg", answ.Message, "type", s.name)

			// case <-time.After(3 * time.Second):
			// 	logging.Debug("greet closed by context")
			// 	return

			case <-ctx.Done():
				logging.Debug("send req to usr", "st", "fail", "rsn", "ctx done", "type", s.name)
				return
			}
		}

		select {
		case s.oxf.Reqs() <- oxf.Word{Text: answ.Message}:
			logging.Debug("send req to oxf", "st", "ok", "type", s.name)

		case <-time.After(1 * time.Second):
			logging.Debug("send req to oxf", "st", "fail", "rsn", "timeout", "type", s.name)
			return

		case <-ctx.Done():
			logging.Debug("send req to oxf", "st", "ok", "type", s.name)
			return
		}

		select {
		case entry := <-s.oxf.Resps():
			logging.Debug("get resp from oxf", "st", "ok", "type", s.name)

			if entry.Word == "" {
				// logging.Debug("check", "st", "ok", "type", s.name)
				// s.usr.Reqs() <- model.Message{Message: "Oh, sorry, I don't know the word, maybe you meant:"}

				Actions := []uint64{}
				Rels := ""

				for _, rel := range entry.RelatedWords {

					select {
					case s.uis.Reqs() <- model.Button{
						Tp:    model.MayBeMeantButton,
						Texts: []string{rel.Text},
					}:
						logging.Debug("send req to uis", "st", "ok", "type", s.name)

						select {
						case id_btn := <-s.uis.Resps():
							logging.Debug("get resp from uis", "st", "ok", "type", s.name)
							Actions = append(Actions, uint64(id_btn))

						case <-time.After(1 * time.Second):
							logging.Debug("get resp from uis", "st", "fail", "rsn", "timeout", "type", s.name)
							Rels += "/n" + rel.Text

						case <-ctx.Done():
							logging.Debug("get resp from uis", "st", "ok", "type", s.name)
							return
						}

					case <-time.After(1 * time.Second):
						logging.Debug("send req to uis", "st", "fail", "rsn", "timeout", "type", s.name)
						Rels += "/n" + rel.Text

					case <-ctx.Done():
						logging.Debug("send req to uis", "st", "fail", "rsn", "ctx done", "type", s.name)
						return
					}
				}

				text := "Oh, sorry, I don't know the word, maybe you meant:" + Rels
				s.usr.Reqs() <- model.Message{Message: text, Actions: Actions}

				return
			}

			out := ""
			for i, sense := range entry.Senses {
				//TODO string builder
				// out += fmt.Sprintf("%d. %s (%s)\n", i, sense.Def, sense.Usage)

				// for _, example := range sense.Examples {
				// 	if example.Usage == "" {
				// 		out += fmt.Sprintf(" - (%s)\n", example.Example)
				// 	} else {
				// 	}
				// }

				out = fmt.Sprintf("%d. <u><strong>%s</strong></u> %s (%s)\n", i+1, entry.Word, sense.Def, sense.Usage)

				for _, example := range sense.Examples {
					if example.Usage == "" {
						out += fmt.Sprintf(" - (%s)\n", example.Example)
					} else {
						out += fmt.Sprintf("<u>%s</u> - (%s)\n", example.Usage, example.Example)
					}
				}

				metaMp := map[string]string{
					"word":  entry.Word,
					"def":   sense.Def,
					"usage": sense.Usage,
					"link":  entry.Link,
					"pos":   strconv.FormatInt(int64(sense.Pos), 10),
					//TODO this hook as button to get req to
					// database should know userId but there is no dialog implementation right now in uis
					// + looks like we should create a new dialog for EVERY button not for user,
					// EVERY button that means we lost info about user Time To Think
				}

				data, err := json.Marshal(metaMp)
				if err != nil {
					logging.Debug("create json for uis btn", "st", "fail", "err", err, "type", s.name)
					s.usr.Reqs() <- model.Message{Message: out}
					return
				}

				select {
				case s.uis.Reqs() <- model.Button{
					Tp:    model.AddToDictButton,
					Texts: []string{"Add to the dict", "Added"},
					Meta:  data,
				}:
					logging.Debug("send req to uis", "st", "ok", "type", s.name)

					select {
					case id_btn := <-s.uis.Resps():
						logging.Debug("get resp from uis", "st", "ok", "type", s.name)
						s.usr.Reqs() <- model.Message{Message: out, Actions: []uint64{uint64(id_btn)}}

					case <-time.After(1 * time.Second):
						logging.Debug("get resp from uis", "st", "fail", "rsn", "timeout", "type", s.name)
						s.usr.Reqs() <- model.Message{Message: out}

					case <-ctx.Done():
						logging.Debug("get resp from uis", "st", "ok", "type", s.name)
						return
					}

				case <-time.After(1 * time.Second):
					logging.Debug("send req to uis", "st", "fail", "rsn", "timeout", "type", s.name)
					s.usr.Reqs() <- model.Message{Message: out}

				case <-ctx.Done():
					logging.Debug("send req to uis", "st", "fail", "rsn", "ctx done", "type", s.name)
					return
				}

			}

		// case <-time.After(3 * time.Second):
		// 	logging.Debug("greet closed by context")
		// 	return

		case <-ctx.Done():
			logging.Debug("get resp from oxf", "st", "fail", "type", s.name)
			return
		}

	// case <-time.After(3 * time.Second):
	// 	logging.Debug("greet closed by context")
	// 	return

	case <-ctx.Done():
		logging.Debug("greet closed by context")
		return
	}
	logging.Debug("greet finished")
}
