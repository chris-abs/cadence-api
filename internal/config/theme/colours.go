package theme

type TailwindColour string

const (
    ColourSlate   TailwindColour = "slate"
    ColourGray    TailwindColour = "gray"
    ColourZinc    TailwindColour = "zinc"
    ColourNeutral TailwindColour = "neutral"
    ColourStone   TailwindColour = "stone"
    ColourRed     TailwindColour = "red"
    ColourOrange  TailwindColour = "orange"
    ColourAmber   TailwindColour = "amber"
    ColourYellow  TailwindColour = "yellow"
    ColourLime    TailwindColour = "lime"
    ColourGreen   TailwindColour = "green"
    ColourEmerald TailwindColour = "emerald"
    ColourTeal    TailwindColour = "teal"
    ColourCyan    TailwindColour = "cyan"
    ColourSky     TailwindColour = "sky"
    ColourBlue    TailwindColour = "blue"
    ColourIndigo  TailwindColour = "indigo"
    ColourViolet  TailwindColour = "violet"
    ColourPurple  TailwindColour = "purple"
    ColourFuchsia TailwindColour = "fuchsia"
    ColourPink    TailwindColour = "pink"
    ColourRose    TailwindColour = "rose"
)

var validColours = map[TailwindColour]bool{
    ColourSlate:   true,
    ColourGray:    true,
    ColourZinc:    true,
    ColourNeutral: true,
    ColourStone:   true,
    ColourRed:     true,
    ColourOrange:  true,
    ColourAmber:   true,
    ColourYellow:  true,
    ColourLime:    true,
    ColourGreen:   true,
    ColourEmerald: true,
    ColourTeal:    true,
    ColourCyan:    true,
    ColourSky:     true,
    ColourBlue:    true,
    ColourIndigo:  true,
    ColourViolet:  true,
    ColourPurple:  true,
    ColourFuchsia: true,
    ColourPink:    true,
    ColourRose:    true,
}

func IsValid(colour string) bool {
    _, valid := validColours[TailwindColour(colour)]
    return valid
}

func GetAllColours() []string {
    colours := make([]string, 0, len(validColours))
    for colour := range validColours {
        colours = append(colours, string(colour))
    }
    return colours
}

func (t TailwindColour) String() string {
    return string(t)
}