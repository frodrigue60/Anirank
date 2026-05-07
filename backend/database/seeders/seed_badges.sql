-- AniRank Badges Seeder
-- This script clears existing badges and populates the hierarchy for XP, Ratings, and Comments.

-- 1. Clean existing data (CASCADE ensures badge_user links are also removed)
TRUNCATE TABLE badge_user CASCADE;
DELETE FROM badges;

-- 2. XP Progression Badges (6 Tiers x 4 Divisions = 24 Badges)

-- Tier: Bronze
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Bronze IV', 'Reached user level 1', true, 'badges/placeholders/xp_bronze_4.webp', true, 'level', 1, gen_random_uuid(), NOW(), NOW()),
('Bronze III', 'Reached user level 5', true, 'badges/placeholders/xp_bronze_3.webp', true, 'level', 5, gen_random_uuid(), NOW(), NOW()),
('Bronze II', 'Reached user level 9', true, 'badges/placeholders/xp_bronze_2.webp', true, 'level', 9, gen_random_uuid(), NOW(), NOW()),
('Bronze I', 'Reached user level 13', true, 'badges/placeholders/xp_bronze_1.webp', true, 'level', 13, gen_random_uuid(), NOW(), NOW());

-- Tier: Silver
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Silver IV', 'Reached user level 17', true, 'badges/placeholders/xp_silver_4.webp', true, 'level', 17, gen_random_uuid(), NOW(), NOW()),
('Silver III', 'Reached user level 21', true, 'badges/placeholders/xp_silver_3.webp', true, 'level', 21, gen_random_uuid(), NOW(), NOW()),
('Silver II', 'Reached user level 25', true, 'badges/placeholders/xp_silver_2.webp', true, 'level', 25, gen_random_uuid(), NOW(), NOW()),
('Silver I', 'Reached user level 29', true, 'badges/placeholders/xp_silver_1.webp', true, 'level', 29, gen_random_uuid(), NOW(), NOW());

-- Tier: Gold
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Gold IV', 'Reached user level 33', true, 'badges/placeholders/xp_gold_4.webp', true, 'level', 33, gen_random_uuid(), NOW(), NOW()),
('Gold III', 'Reached user level 37', true, 'badges/placeholders/xp_gold_3.webp', true, 'level', 37, gen_random_uuid(), NOW(), NOW()),
('Gold II', 'Reached user level 41', true, 'badges/placeholders/xp_gold_2.webp', true, 'level', 41, gen_random_uuid(), NOW(), NOW()),
('Gold I', 'Reached user level 45', true, 'badges/placeholders/xp_gold_1.webp', true, 'level', 45, gen_random_uuid(), NOW(), NOW());

-- Tier: Platinum
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Platinum IV', 'Reached user level 49', true, 'badges/placeholders/xp_platinum_4.webp', true, 'level', 49, gen_random_uuid(), NOW(), NOW()),
('Platinum III', 'Reached user level 53', true, 'badges/placeholders/xp_platinum_3.webp', true, 'level', 53, gen_random_uuid(), NOW(), NOW()),
('Platinum II', 'Reached user level 57', true, 'badges/placeholders/xp_platinum_2.webp', true, 'level', 57, gen_random_uuid(), NOW(), NOW()),
('Platinum I', 'Reached user level 61', true, 'badges/placeholders/xp_platinum_1.webp', true, 'level', 61, gen_random_uuid(), NOW(), NOW());

-- Tier: Emerald
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Emerald IV', 'Reached user level 65', true, 'badges/placeholders/xp_emerald_4.webp', true, 'level', 65, gen_random_uuid(), NOW(), NOW()),
('Emerald III', 'Reached user level 69', true, 'badges/placeholders/xp_emerald_3.webp', true, 'level', 69, gen_random_uuid(), NOW(), NOW()),
('Emerald II', 'Reached user level 73', true, 'badges/placeholders/xp_emerald_2.webp', true, 'level', 73, gen_random_uuid(), NOW(), NOW()),
('Emerald I', 'Reached user level 77', true, 'badges/placeholders/xp_emerald_1.webp', true, 'level', 77, gen_random_uuid(), NOW(), NOW());

-- Tier: Diamond
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Diamond IV', 'Reached user level 81', true, 'badges/placeholders/xp_diamond_4.webp', true, 'level', 81, gen_random_uuid(), NOW(), NOW()),
('Diamond III', 'Reached user level 86', true, 'badges/placeholders/xp_diamond_3.webp', true, 'level', 86, gen_random_uuid(), NOW(), NOW()),
('Diamond II', 'Reached user level 91', true, 'badges/placeholders/xp_diamond_2.webp', true, 'level', 91, gen_random_uuid(), NOW(), NOW()),
('Diamond I', 'Reached user level 96', true, 'badges/placeholders/xp_diamond_1.webp', true, 'level', 96, gen_random_uuid(), NOW(), NOW());

-- 3. Rating Milestones (9 Badges)
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Initiate Listener', 'Rated 10 themes', true, 'badges/placeholders/rating_1.webp', true, 'ratings', 10, gen_random_uuid(), NOW(), NOW()),
('Melody Seeker', 'Rated 50 themes', true, 'badges/placeholders/rating_2.webp', true, 'ratings', 50, gen_random_uuid(), NOW(), NOW()),
('Active Auditor', 'Rated 100 themes', true, 'badges/placeholders/rating_3.webp', true, 'ratings', 100, gen_random_uuid(), NOW(), NOW()),
('Soundtrack Specialist', 'Rated 250 themes', true, 'badges/placeholders/rating_4.webp', true, 'ratings', 250, gen_random_uuid(), NOW(), NOW()),
('Theme Collector', 'Rated 500 themes', true, 'badges/placeholders/rating_5.webp', true, 'ratings', 500, gen_random_uuid(), NOW(), NOW()),
('Rhythm Scholar', 'Rated 1,000 themes', true, 'badges/placeholders/rating_6.webp', true, 'ratings', 1000, gen_random_uuid(), NOW(), NOW()),
('Archive Guardian', 'Rated 2,500 themes', true, 'badges/placeholders/rating_7.webp', true, 'ratings', 2500, gen_random_uuid(), NOW(), NOW()),
('Symphony Master', 'Rated 5,000 themes', true, 'badges/placeholders/rating_8.webp', true, 'ratings', 5000, gen_random_uuid(), NOW(), NOW()),
('AniRank Grand Maestro', 'Rated 10,000 themes', true, 'badges/placeholders/rating_9.webp', true, 'ratings', 10000, gen_random_uuid(), NOW(), NOW());

-- 4. Commenting Milestones (9 Badges)
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('First Word', 'Posted your first comment', true, 'badges/placeholders/comment_1.webp', true, 'comments', 1, gen_random_uuid(), NOW(), NOW()),
('Casual Commenter', 'Posted 10 comments', true, 'badges/placeholders/comment_2.webp', true, 'comments', 10, gen_random_uuid(), NOW(), NOW()),
('Active Talker', 'Posted 25 comments', true, 'badges/placeholders/comment_3.webp', true, 'comments', 25, gen_random_uuid(), NOW(), NOW()),
('Debater', 'Posted 50 comments', true, 'badges/placeholders/comment_4.webp', true, 'comments', 50, gen_random_uuid(), NOW(), NOW()),
('Community Pillar', 'Posted 100 comments', true, 'badges/placeholders/comment_5.webp', true, 'comments', 100, gen_random_uuid(), NOW(), NOW()),
('Opinion Leader', 'Posted 250 comments', true, 'badges/placeholders/comment_6.webp', true, 'comments', 250, gen_random_uuid(), NOW(), NOW()),
('Respected Orator', 'Posted 500 comments', true, 'badges/placeholders/comment_7.webp', true, 'comments', 500, gen_random_uuid(), NOW(), NOW()),
('Legendary Critic', 'Posted 1,000 comments', true, 'badges/placeholders/comment_8.webp', true, 'comments', 1000, gen_random_uuid(), NOW(), NOW()),
('Universal Oracle', 'Posted 2,500 comments', true, 'badges/placeholders/comment_9.webp', true, 'comments', 2500, gen_random_uuid(), NOW(), NOW());

-- 5. Special Badges
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('AniFan', 'Connected your AniList account', true, 'badges/placeholders/anifan.webp', true, 'anilist', 1, gen_random_uuid(), NOW(), NOW());
