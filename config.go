package spinlog

type Config struct {
	LogDir   LogDir `json:"log_dir"`
	MaxCount int    `json:"max_count"`
	MaxSize  int64  `json:"max_size"`
	SetPerm  bool   `json:"set_perm"`
	Perm     int    `json:"perm"`
	SetOwner bool   `json:"set_owner"`
	GID      int    `json:"gid"`
	UID      int    `json:"uid"`
}

type LineConfig struct {
	Config
	MaxLineSize int `json:"max_line_size"`
}
