package modrinth

type ProjectType string

const (
	Mod          ProjectType = "mod"
	Modpack      ProjectType = "modpack"
	Plugin       ProjectType = "plugin"
	Resourcepack ProjectType = "resourcepack"
)

type ProjectStatus string

const (
	Approved   ProjectStatus = "approved"
	Rejected   ProjectStatus = "rejected"
	Draft      ProjectStatus = "draft"
	Unlisted   ProjectStatus = "unlisted"
	Archived   ProjectStatus = "archived"
	Processing ProjectStatus = "processing"
	Unknown    ProjectStatus = "unknown"
)

type EnvSupport string

const (
	Required    EnvSupport = "required"
	Optional    EnvSupport = "optional"
	Unsupported EnvSupport = "unsupported"
)

type ReleaseChannel string

const (
	Release ReleaseChannel = "release"
	Beta    ReleaseChannel = "beta"
	Alpha   ReleaseChannel = "alpha"
)

type HashAlgo string

const (
	Sha1   HashAlgo = "sha1"
	Sha512 HashAlgo = "sha512"
)

type LabrinthInfo struct {
	About         string `json:"about"`
	Documentation string `json:"documentation"`
	Name          string `json:"name"`
	Version       string `json:"version"`
}

type Project struct {
	Slug                 string           `json:"slug"`
	Title                string           `json:"title"`
	Description          string           `json:"description"`
	Categories           []string         `json:"categories"`
	ClientSide           EnvSupport       `json:"client_side"`
	ServerSide           EnvSupport       `json:"server_side"`
	Body                 string           `json:"body"`
	AdditionalCategories []string         `json:"additional_categories"`
	IssuesUrl            string           `json:"issues_url"`
	SourceUrl            string           `json:"source_url"`
	WikiUrl              string           `json:"wiki_url"`
	DiscordUrl           string           `json:"discord_url"`
	DonationUrls         []DonationUrl    `json:"donation_urls"`
	ProjectType          ProjectType      `json:"project_type"`
	Downloads            int              `json:"downloads"`
	IconUrl              string           `json:"icon_url"`
	Id                   string           `json:"id"`
	Team                 string           `json:"team"`
	ModeratorMessage     ModeratorMessage `json:"moderator_message"`
	Published            string           `json:"published"`
	Updated              string           `json:"updated"`
	Approved             string           `json:"approved"`
	Followers            int              `json:"followers"`
	Status               ProjectStatus    `json:"status"`
	License              License          `json:"license"`
	Versions             []string         `json:"versions"`
	Gallery              []GalleryItem    `json:"gallery"`
}

type DonationUrl struct {
	Id       string `json:"id"`
	Platform string `json:"platform"`
	Url      string `json:"url"`
}

type ModeratorMessage struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

type License struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GalleryItem struct {
	Url         string `json:"url"`
	Featured    bool   `json:"featured"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Created     string `json:"created"`
}

type Version struct {
	Name          string         `json:"name"`
	VersionNumber string         `json:"version_number"`
	Changelog     string         `json:"changelog"`
	Dependencies  []Dependency   `json:"dependencies"`
	GameVersions  []string       `json:"game_versions"`
	VersionType   ReleaseChannel `json:"version_type"`
	Loaders       []string       `json:"loaders"`
	Featured      bool           `json:"featured"`
	Id            string         `json:"id"`
	ProjectId     string         `json:"project_id"`
	AuthorId      string         `json:"author_id"`
	DatePublished string         `json:"date_published"`
	Downloads     int            `json:"downloads"`
	Files         []File         `json:"files"`
}

type Dependency struct {
	VersionId      string `json:"version_id"`
	ProjectId      string `json:"project_id"`
	FileName       string `json:"file_name"`
	DependencyType string `json:"dependency_type"`
}

type File struct {
	Hashes   Hashes `json:"hashes"`
	Url      string `json:"url"`
	Filename string `json:"filename"`
	Primary  bool   `json:"primary"`
	Size     int    `json:"size"`
}

type Hashes struct {
	Sha512 HashAlgo `json:"sha512"`
	Sha1   HashAlgo `json:"sha1"`
}

type Dependencies struct {
	Projects []Project `json:"projects"`
	Versions []Version `json:"versions"`
}

type CheckResponse struct {
	Id string `json:"id"`
}
