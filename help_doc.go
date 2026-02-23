package mflags

import "encoding/json"

// ChoiceProvider is an interface for Value types that expose a set of valid choices.
// The existing choiceValue type already satisfies this interface.
type ChoiceProvider interface {
	Choices() []string
}

// HelpDocument is the top-level structure for JSON help output from a Dispatcher.
type HelpDocument struct {
	Name     string       `json:"name"`
	Commands []CommandDoc `json:"commands"`
}

// CommandDoc describes a single command in the help document.
type CommandDoc struct {
	Path            string          `json:"path"`
	Usage           string          `json:"usage"`
	Description     string          `json:"description,omitempty"`
	RequiredFeature string          `json:"requiredFeature,omitempty"`
	Flags           []FlagDoc       `json:"flags"`
	FlagGroups      []string        `json:"flagGroups"`
	PositionalArgs  []PositionalDoc `json:"positionalArgs,omitempty"`
	AcceptsRestArgs bool            `json:"acceptsRestArgs,omitempty"`
	Examples        []ExampleDoc    `json:"examples,omitempty"`
	Subcommands     []CommandDoc    `json:"subcommands,omitempty"`
}

// ExampleDoc describes a usage example in the help document.
type ExampleDoc struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// FlagDoc describes a single flag in the help document.
type FlagDoc struct {
	Name    string   `json:"name"`
	Short   string   `json:"short"`
	Type    string   `json:"type"`
	Default string   `json:"default,omitempty"`
	Usage   string   `json:"usage"`
	Group   string   `json:"group,omitempty"`
	IsBool  bool     `json:"isBool,omitempty"`
	Choices []string `json:"choices,omitempty"`
}

// PositionalDoc describes a positional argument in the help document.
type PositionalDoc struct {
	Name  string `json:"name"`
	Usage string `json:"usage"`
	Type  string `json:"type"`
}

// FlagSetDoc describes a standalone FlagSet's help information.
type FlagSetDoc struct {
	Name            string          `json:"name"`
	Flags           []FlagDoc       `json:"flags"`
	FlagGroups      []string        `json:"flagGroups"`
	PositionalArgs  []PositionalDoc `json:"positionalArgs"`
	AcceptsRestArgs bool            `json:"acceptsRestArgs"`
}

// HelpDoc generates a structured help document for a FlagSet.
func (f *FlagSet) HelpDoc() *FlagSetDoc {
	doc := &FlagSetDoc{
		Name:            f.name,
		Flags:           []FlagDoc{},
		FlagGroups:      []string{},
		PositionalArgs:  []PositionalDoc{},
		AcceptsRestArgs: f.restField != nil,
	}

	f.VisitAll(func(flag *Flag) {
		fd := FlagDoc{
			Name:    flag.Name,
			Type:    flag.Value.Type(),
			Default: flag.DefValue,
			Usage:   flag.Usage,
			Group:   flag.Group,
			IsBool:  flag.Value.IsBool(),
			Choices: []string{},
		}
		if flag.Short != 0 {
			fd.Short = string(flag.Short)
		}
		if cp, ok := flag.Value.(ChoiceProvider); ok {
			fd.Choices = cp.Choices()
		}
		doc.Flags = append(doc.Flags, fd)
	})

	if len(f.groupOrder) > 0 {
		doc.FlagGroups = append(doc.FlagGroups, f.groupOrder...)
	}

	for _, pf := range f.GetPositionalFields() {
		doc.PositionalArgs = append(doc.PositionalArgs, PositionalDoc{
			Name:  pf.Name,
			Usage: pf.Usage,
			Type:  pf.Type.String(),
		})
	}

	return doc
}

// HelpJSON returns the FlagSet's help document as indented JSON.
func (f *FlagSet) HelpJSON() ([]byte, error) {
	return json.MarshalIndent(f.HelpDoc(), "", "  ")
}

// HelpDoc generates a structured help document for the entire Dispatcher.
func (d *Dispatcher) HelpDoc() *HelpDocument {
	doc := &HelpDocument{
		Name:     d.name,
		Commands: d.buildCommandDocs(""),
	}
	return doc
}

// HelpJSON returns the Dispatcher's help document as indented JSON.
func (d *Dispatcher) HelpJSON() ([]byte, error) {
	return json.MarshalIndent(d.HelpDoc(), "", "  ")
}

// buildCommandDocs recursively builds CommandDoc entries for direct children of parentPath.
func (d *Dispatcher) buildCommandDocs(parentPath string) []CommandDoc {
	children := d.getDirectChildren(parentPath)
	docs := make([]CommandDoc, 0, len(children))

	for _, child := range children {
		cmd := CommandDoc{
			Path:            child.Path,
			Usage:           child.Usage,
			Flags:           []FlagDoc{},
			FlagGroups:      []string{},
			PositionalArgs:  []PositionalDoc{},
			AcceptsRestArgs: false,
			Subcommands:     []CommandDoc{},
		}

		if child.IsEntry {
			entry := d.commands[child.Path]
			if entry != nil && entry.Command != nil {
				fs := entry.Command.FlagSet()
				if fs != nil {
					fsDoc := fs.HelpDoc()
					cmd.Flags = fsDoc.Flags
					cmd.FlagGroups = fsDoc.FlagGroups
					cmd.PositionalArgs = fsDoc.PositionalArgs
					cmd.AcceptsRestArgs = fsDoc.AcceptsRestArgs
				}
				if ep, ok := entry.Command.(ExampleProvider); ok {
					for _, ex := range ep.Examples() {
						cmd.Examples = append(cmd.Examples, ExampleDoc{
							Name: ex.Name,
							Body: ex.Body,
						})
					}
				}
				if rfp, ok := entry.Command.(RequiredFeatureProvider); ok {
					cmd.RequiredFeature = rfp.RequiredFeature()
				}
				if dp, ok := entry.Command.(DescriptionProvider); ok {
					cmd.Description = dp.Description()
				}
			}
		}

		cmd.Subcommands = d.buildCommandDocs(child.Path)

		docs = append(docs, cmd)
	}

	return docs
}
