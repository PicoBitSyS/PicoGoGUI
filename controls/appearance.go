package controls

// Appearance describes visual properties shared by the basic controls.
// Zero values keep the active theme defaults.
type Appearance struct {
	FontFamily   string  `json:"fontFamily,omitempty"`
	FontSize     int     `json:"fontSize,omitempty"`
	Color        string  `json:"color,omitempty"`
	Background   string  `json:"background,omitempty"`
	Bold         bool    `json:"bold,omitempty"`
	Italic       bool    `json:"italic,omitempty"`
	Underline    bool    `json:"underline,omitempty"`
	TextAlign    string  `json:"textAlign,omitempty"`
	BorderColor  string  `json:"borderColor,omitempty"`
	BorderWidth  int     `json:"borderWidth,omitempty"`
	BorderRadius int     `json:"borderRadius,omitempty"`
	Opacity      float64 `json:"opacity,omitempty"`
}
