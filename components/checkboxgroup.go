package components

import (
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/util"
	"github.com/jiiihpeeh/recliner/vdom"
)

type CheckboxOption struct {
	ID    string
	Label string
	Value bool
}

func CheckboxGroup(props any) vdom.Node {
	options, _ := util.GetProp[[]CheckboxOption](props, "options")
	onChange, _ := util.GetProp[func([]CheckboxOption)](props, "onChange")
	label, _ := util.GetProp[string](props, "label")

	checkedCount := 0
	for _, opt := range options {
		if opt.Value {
			checkedCount++
		}
	}

	var triState *bool
	if len(options) > 0 {
		if checkedCount == len(options) {
			triState = func() *bool { v := true; return &v }()
		} else if checkedCount == 0 {
			triState = func() *bool { v := false; return &v }()
		} else {
			triState = nil
		}
	}

	handleGroupChange := func(newValue bool) {
		newOptions := make([]CheckboxOption, len(options))
		for i := range options {
			newOptions[i] = CheckboxOption{
				ID:    options[i].ID,
				Label: options[i].Label,
				Value: newValue,
			}
		}
		if onChange != nil {
			onChange(newOptions)
		}
	}

	renderCheckbox := func(opt CheckboxOption, index int) vdom.Node {
		elementProps := struct {
			Style   vdom.Style
			OnClick func(events.MouseEvent)
		}{
			Style: vdom.Style{},
		}
		elementProps.OnClick = func(e events.MouseEvent) {
			if e.Action == events.MouseActionPress {
				newOptions := make([]CheckboxOption, len(options))
				for i := range options {
					newOptions[i] = options[i]
					if options[i].ID == opt.ID {
						newOptions[i].Value = !options[i].Value
					}
				}
				if onChange != nil {
					onChange(newOptions)
				}
			}
		}

		checkSymbol := "☐"
		if opt.Value {
			checkSymbol = "☑"
		}

		return Box(struct {
			Style    vdom.Style
			Children []vdom.Node
		}{
			Style: vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionRow, Gap: 1},
			Children: []vdom.Node{
				Text(struct {
					Children string
					Style    vdom.Style
					OnClick  func(events.MouseEvent)
				}{
					Children: checkSymbol,
					Style:    vdom.Style{},
					OnClick:  elementProps.OnClick,
				}),
				Text(struct {
					Children string
					Style    vdom.Style
				}{
					Children: opt.Label,
					Style:    vdom.Style{},
				}),
			},
		})
	}

	groupCheckbox := func() vdom.Node {
		checkSymbol := "☐"
		if triState != nil && *triState {
			checkSymbol = "☑"
		} else if triState == nil && len(options) > 0 {
			checkSymbol = "◠"
		}

		elementProps := struct {
			OnClick func(events.MouseEvent)
		}{}
		elementProps.OnClick = func(e events.MouseEvent) {
			if e.Action == events.MouseActionPress {
				if triState == nil {
					handleGroupChange(true)
				} else if *triState {
					handleGroupChange(false)
				} else {
					handleGroupChange(true)
				}
			}
		}

		return Box(struct {
			Style    vdom.Style
			Children []vdom.Node
		}{
			Style: vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionRow, Gap: 1},
			Children: []vdom.Node{
				Text(struct {
					Children string
					Style    vdom.Style
					OnClick  func(events.MouseEvent)
				}{
					Children: checkSymbol,
					Style:    vdom.Style{Bold: true},
					OnClick:  elementProps.OnClick,
				}),
				Text(struct {
					Children string
					Style    vdom.Style
				}{
					Children: label,
					Style:    vdom.Style{Bold: true},
				}),
				Text(struct {
					Children string
					Style    vdom.Style
				}{
					Children: " (" + string(rune('0'+checkedCount)) + "/" + string(rune('0'+len(options))) + ")",
					Style:    vdom.Style{Foreground: "gray"},
				}),
			},
		})
	}

	children := []vdom.Node{}
	if label != "" {
		children = append(children, groupCheckbox())
		children = append(children, Newline(1))
	}

	for i, opt := range options {
		children = append(children, Text(struct {
			Children string
			Style    vdom.Style
		}{
			Children: "  ",
			Style:    vdom.Style{},
		}))
		children = append(children, renderCheckbox(opt, i))
		children = append(children, Newline(1))
	}

	return Box(struct {
		Style    vdom.Style
		Children []vdom.Node
	}{
		Style:    vdom.Style{Display: vdom.DisplayFlex, FlexDirection: vdom.FlexDirectionColumn},
		Children: children,
	})
}
