-- +migrate Up
INSERT INTO permissions (slug, name, description) 
VALUES ('partners.manage', 'Manage Partners', 'Ability to create, edit and delete partners/communities banners')
ON CONFLICT (slug) DO NOTHING;

-- Assign to owner, admin and editor roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.slug IN ('owner', 'admin', 'editor') AND p.slug = 'partners.manage'
ON CONFLICT DO NOTHING;

-- +migrate Down
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE slug = 'partners.manage');
DELETE FROM permissions WHERE slug = 'partners.manage';
