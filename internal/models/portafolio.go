package models

// Los campos *EN guardan la versión en inglés; si están vacíos se muestra el español.

type Experience struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Place         string `json:"place"`
	Description   string `json:"description"`
	StartYear     int    `json:"start_year"`
	EndYear       int    `json:"end_year"`
	TitleEN       string `json:"title_en,omitempty"`
	PlaceEN       string `json:"place_en,omitempty"`
	DescriptionEN string `json:"description_en,omitempty"`
}

type Education struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Place     string `json:"place"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
	TitleEN   string `json:"title_en,omitempty"`
	PlaceEN   string `json:"place_en,omitempty"`
}

type Project struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	ProjectURL    string `json:"project_url"`
	TitleEN       string `json:"title_en,omitempty"`
	DescriptionEN string `json:"description_en,omitempty"`
}

// Localize* devuelven la versión en el idioma pedido (con respaldo en español).

func (e Experience) Localize(lang string) Experience {
	if lang == "en" {
		e.Title, e.Place, e.Description = or(e.TitleEN, e.Title), or(e.PlaceEN, e.Place), or(e.DescriptionEN, e.Description)
	}
	return e
}

func (e Education) Localize(lang string) Education {
	if lang == "en" {
		e.Title, e.Place = or(e.TitleEN, e.Title), or(e.PlaceEN, e.Place)
	}
	return e
}

func (p Project) Localize(lang string) Project {
	if lang == "en" {
		p.Title, p.Description = or(p.TitleEN, p.Title), or(p.DescriptionEN, p.Description)
	}
	return p
}

func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
