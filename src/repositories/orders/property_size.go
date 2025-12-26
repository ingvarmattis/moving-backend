package orders

type PropertySize int8

const (
	PropertySizeUnknown PropertySize = iota
	PropertySizeStudio
	PropertySize1Bedroom
	PropertySize2Bedrooms
	PropertySize3Bedrooms
	PropertySize4PlusBedrooms
	PropertySizeCommercial
)

func (p PropertySize) String() string {
	switch p {
	case PropertySizeStudio:
		return "studio"
	case PropertySize1Bedroom:
		return "1_bedroom"
	case PropertySize2Bedrooms:
		return "2_bedrooms"
	case PropertySize3Bedrooms:
		return "3_bedrooms"
	case PropertySize4PlusBedrooms:
		return "4_plus_bedrooms"
	case PropertySizeCommercial:
		return "commercial"
	default:
		return "unknown"
	}
}

func NewPropertySize(s string) PropertySize {
	switch s {
	case "studio":
		return PropertySizeStudio
	case "1_bedroom":
		return PropertySize1Bedroom
	case "2_bedrooms":
		return PropertySize2Bedrooms
	case "3_bedrooms":
		return PropertySize3Bedrooms
	case "4_plus_bedrooms":
		return PropertySize4PlusBedrooms
	case "commercial":
		return PropertySizeCommercial
	default:
		return PropertySizeUnknown
	}
}
