package flacmeta

type VorbisCommentBlock struct {
	Vendor   string
	Comments map[string]string
}

func NewVorbisCommentBlock() *VorbisCommentBlock {
	return &VorbisCommentBlock{
		Vendor:   "gcottom-flacmeta",
		Comments: make(map[string]string),
	}
}

func (v *VorbisCommentBlock) GetVendor() string {
	return v.Vendor
}

func (v *VorbisCommentBlock) GetComments() map[string]string {
	return v.Comments
}

func (v *VorbisCommentBlock) GetComment(key string) string {
	return v.Comments[key]
}

func (v *VorbisCommentBlock) SetVendor(vendor string) {
	v.Vendor = vendor
}

func (v *VorbisCommentBlock) SetComment(key string, value string) error {
	for _, c := range key {
		if c == '=' || c < 0x20 || c > 0x7D {
			return &ErrInvalidFieldName{Field: key}
		}
	}
	v.Comments[key] = value
	return nil
}

func (v *VorbisCommentBlock) DeleteComment(key string) {
	delete(v.Comments, key)
}
