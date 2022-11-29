package iptables

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plockc/gateway/funcs"
	"github.com/plockc/gateway/resource"
)

var RuleIdRegex = regexp.MustCompile(`.*gw-dt\[([0-9a-f]+)]: (.*)`)

type Rule struct {
	Id          uint32 `json:"-"`
	Chain       `json:"-"`
	Target      string     `json:"target"`
	Start       *time.Time `json:"start"`
	End         *time.Time `json:"end"`
	MatchSetSrc string     `json:"matchSetSrc"`
	Comment     string     `json:"comment"`
}

func NewRule(c Chain) Rule {
	return Rule{
		Id:     rand.Uint32(),
		Chain:  c,
		Target: "RETURN",
	}
}

func (r Rule) RuleResource() RuleRes {
	return RuleRes{Rule: r}
}

// TODO: sanitize the comment
func (r Rule) CoreArgs() []string {
	args := []string{r.Chain.Name, "-t", r.Table.Name}
	match := []string{}
	if len(r.MatchSetSrc) > 0 {
		match = append(match, []string{"-m", "set", "--match-set", r.MatchSetSrc, "src"}...)
	}
	return append(args, match...)
}

func (r Rule) Args() []string {
	comment := []string{"-m", "comment", "--comment", fmt.Sprintf("gw-dt[%0.8x]: %s", r.Id, r.Comment)}
	return append(r.CoreArgs(), append(comment, "-j", r.Target)...)
}

func (r Rule) String() string {
	return r.Table.String() + ":rule[" + strings.Join(r.Args(), " ") + " ]"
}

var _ resource.Resource = RuleRes{}

type RuleRes struct {
	Rule
	resource.FailUnimplementedMethods
}

func (r RuleRes) Id() string {
	return fmt.Sprintf("%0.8x", r.Rule.Id)
}

func (r RuleRes) Create() error {
	return r.Runner().Batch(append([]string{"iptables", "-A"}, r.Args()...))
}

func (r RuleRes) Delete() error {
	return r.Runner().Batch(append([]string{"iptables", "-D"}, r.Args()...))
}

func (r RuleRes) List() ([]string, error) {
	res, err := r.Runner().Exec(append([]string{"iptables", "-v", "-L"}, r.CoreArgs()...))
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, l := range strings.Split(res.Out, "\n") {
		matches := RuleIdRegex.FindStringSubmatch(l)
		if len(matches) <= 1 {
			continue
		}
		ids = append(ids, matches[1])
	}
	return ids, nil
}

func (r RuleRes) Clear() error {
	ruleIds, err := r.List()
	if err != nil {
		return err
	}
	for _, ruleId := range ruleIds {
		r.Rule.Id, err = ParseRuleId(ruleId)
		if err != nil {
			return err
		}
		err = r.Load()
		if err != nil {
			return err
		}
		err = r.Delete()
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseRuleId(s string) (uint32, error) {
	rId, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(rId), nil
}

func (r *RuleRes) Load() error {
	res, err := r.Runner().Exec([]string{"iptables-save", "-t", r.Table.Name})
	if err != nil {
		return err
	}
	// remove rules not in the chain or not managed by this program
	rules := funcs.Keep(strings.Split(res.Out, "\n"), func(s string) bool {
		if !strings.HasPrefix(s, "-A "+r.Chain.Name+" ") {
			return false
		}
		matches := RuleIdRegex.FindStringSubmatch(s)
		return len(matches) >= 2
	})
	if len(rules) < 1 {
		return fmt.Errorf("expected a matching rule, got %v", rules)
	}

	ruleSpec := strings.Split(rules[0], " ")
	i := 2
	for i < len(ruleSpec) {
		switch ruleSpec[i] {
		case "-m":
			switch ruleSpec[i+1] {
			case "set":
				if ruleSpec[i+2] != "--match-set" {
					return fmt.Errorf("failed to find match-set arg for -m set: %s", ruleSpec)
				}
				if ruleSpec[i+4] != "src" {
					return fmt.Errorf("only supporting src for match-set: %s", ruleSpec)
				}
				r.MatchSetSrc = ruleSpec[i+3]
				i += 5
			case "comment":
				if ruleSpec[i+2] != "--comment" {
					return fmt.Errorf("failed to find comment arg for -m comment: %s", ruleSpec)
				}
				i += 3
				commentParts := []string{}
				for {
					if !strings.HasSuffix(ruleSpec[i], `"`) {
						commentParts = append(commentParts, ruleSpec[i])
						i++
					} else {
						commentParts = append(commentParts, ruleSpec[i])
						i++
						break
					}
				}
				commentMatch := RuleIdRegex.FindStringSubmatch(strings.Trim(strings.Join(commentParts, " "), `"`))
				if len(commentMatch) != 3 {
					return fmt.Errorf("failed to process Id from comment: %v", commentParts)
				}
				r.Rule.Id, err = ParseRuleId(commentMatch[1])
				if err != nil {
					return err
				}
				r.Comment = commentMatch[2]
			}
		case "-j":
			r.Target = ruleSpec[i+1]
			i += 2
		default:
			return fmt.Errorf("failed to parse at index %d: %s", i, ruleSpec[i])
		}
	}
	return nil
}
