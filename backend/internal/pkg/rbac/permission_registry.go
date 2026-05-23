package rbac

import "fmt"

// PermissionDef represents a single permission to be synced with the database.
type PermissionDef struct {
	Name        string
	Slug        string
	Description string
}

// RolePermissionMatrix maps role slugs to the permission slugs they should have.
type RolePermissionMatrix map[string][]string

// --- Entity & Action Definitions ---

// CRUDEntities are top-level entities that follow the standard create/edit/delete pattern.
var CRUDEntities = []string{
	"anime",
	"song",
	"artist",
}

// TaxonomyEntities are taxonomy sub-domains that also follow create/edit/delete.
var TaxonomyEntities = []string{
	"years",
	"seasons",
	"formats",
	"genres",
	"studios",
	"producers",
}

// CRUDActions are the standard actions generated for each entity.
var CRUDActions = []string{"create", "edit", "delete"}

// ManagePermission defines a standalone "manage" permission.
type ManagePermission struct {
	Entity      string // e.g. "reports"
	Name        string // e.g. "Manage Reports"
	Description string
}

// ManagePermissions are special permissions that don't follow the entity.action CRUD pattern.
var ManagePermissions = []ManagePermission{
	{Entity: "reports", Name: "Manage Reports", Description: "Allows resolving moderation/content reports (songs, comments, users)"},
	{Entity: "users", Name: "Manage Users", Description: "Allows listing and editing user profiles"},
	{Entity: "permissions", Name: "Manage Permissions", Description: "Allows modifying role-permission mappings"},
	{Entity: "tournament", Name: "Manage Tournaments", Description: "Allows creating, seeding and managing tournaments"},
	{Entity: "announcement", Name: "Manage Announcements", Description: "Allows creating and managing platform announcements"},
	{Entity: "badge", Name: "Manage Badges", Description: "Allows creating and managing system badges"},
	{Entity: "webhooks", Name: "Manage Webhooks", Description: "Allows creating and managing webhook integrations"},
	{Entity: "partners", Name: "Manage Partners", Description: "Allows creating and managing partner/community banners"},
}

// --- Permission Generation ---

// BuildPermissionRegistry generates the full list of PermissionDefs from the matrix.
func BuildPermissionRegistry() []PermissionDef {
	var perms []PermissionDef

	// 1. CRUD Permissions: entity.action
	for _, entity := range CRUDEntities {
		for _, action := range CRUDActions {
			perms = append(perms, PermissionDef{
				Name:        fmt.Sprintf("%s %s", actionLabel(action), entityLabel(entity)),
				Slug:        fmt.Sprintf("%s.%s", entity, action),
				Description: fmt.Sprintf("Allows %s %s", actionVerb(action), entity),
			})
		}
	}

	// 2. Taxonomy Permissions: taxonomy.subdomain.action
	for _, sub := range TaxonomyEntities {
		for _, action := range CRUDActions {
			perms = append(perms, PermissionDef{
				Name:        fmt.Sprintf("%s %s", actionLabel(action), entityLabel(sub)),
				Slug:        fmt.Sprintf("taxonomy.%s.%s", sub, action),
				Description: fmt.Sprintf("Allows %s %s taxonomy entries", actionVerb(action), sub),
			})
		}
	}

	// 3. Manage Permissions: entity.manage
	for _, mp := range ManagePermissions {
		perms = append(perms, PermissionDef{
			Name:        mp.Name,
			Slug:        fmt.Sprintf("%s.manage", mp.Entity),
			Description: mp.Description,
		})
	}

	return perms
}

// BuildRoleMatrix generates the role -> permission slug mappings.
// Owner is not included because it has a middleware-level bypass.
func BuildRoleMatrix(allPerms []PermissionDef) RolePermissionMatrix {
	// Collect all slugs for quick lookup
	allSlugs := make(map[string]bool, len(allPerms))
	for _, p := range allPerms {
		allSlugs[p.Slug] = true
	}

	// Admin: gets ALL permissions
	var adminPerms []string
	for _, p := range allPerms {
		adminPerms = append(adminPerms, p.Slug)
	}

	// Editor: gets create + edit (no delete), plus selected manage permissions
	editorManagePerms := map[string]bool{
		"reports.manage":    true,
		"tournament.manage": true,
	}
	var editorPerms []string
	for _, p := range allPerms {
		slug := p.Slug
		// CRUD/Taxonomy: create and edit only (no delete)
		if isEntityAction(slug, "create") || isEntityAction(slug, "edit") {
			editorPerms = append(editorPerms, slug)
			continue
		}
		// Manage permissions: only the allowed ones
		if editorManagePerms[slug] {
			editorPerms = append(editorPerms, slug)
		}
	}

	// Creator: gets only create actions (CRUD + taxonomy)
	var creatorPerms []string
	for _, p := range allPerms {
		if isEntityAction(p.Slug, "create") {
			creatorPerms = append(creatorPerms, p.Slug)
		}
	}

	return RolePermissionMatrix{
		"admin":   adminPerms,
		"editor":  editorPerms,
		"creator": creatorPerms,
	}
}

// --- Helpers ---

// isEntityAction checks if a slug ends with a given action suffix.
// Works for both "entity.action" and "taxonomy.sub.action" patterns.
func isEntityAction(slug, action string) bool {
	suffix := "." + action
	return len(slug) > len(suffix) && slug[len(slug)-len(suffix):] == suffix
}

func actionLabel(action string) string {
	switch action {
	case "create":
		return "Create"
	case "edit":
		return "Edit"
	case "delete":
		return "Delete"
	default:
		return action
	}
}

func entityLabel(entity string) string {
	switch entity {
	case "anime":
		return "Anime"
	case "song":
		return "Song"
	case "artist":
		return "Artist"
	case "years":
		return "Years"
	case "seasons":
		return "Seasons"
	case "formats":
		return "Formats"
	case "genres":
		return "Genres"
	case "studios":
		return "Studios"
	case "producers":
		return "Producers"
	default:
		return entity
	}
}

func actionVerb(action string) string {
	switch action {
	case "create":
		return "creating"
	case "edit":
		return "editing"
	case "delete":
		return "deleting"
	default:
		return action
	}
}
