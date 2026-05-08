-- AniRank Badges Seeder
-- This script clears existing badges and populates the hierarchy for XP, Ratings, and Comments.

-- 1. Clean existing data (CASCADE ensures badge_user links are also removed)
TRUNCATE TABLE badge_user CASCADE;
DELETE FROM badges;

-- 2. XP Progression Badges (16 Titles distributed across 100 levels)
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Newbie', 'Reached user level 1', true, 'badges/placeholders/xp_1.avif', true, 'level', 1, gen_random_uuid(), NOW(), NOW()),
('Rookie', 'Reached user level 5', true, 'badges/placeholders/xp_2.avif', true, 'level', 5, gen_random_uuid(), NOW(), NOW()),
('Apprentice', 'Reached user level 10', true, 'badges/placeholders/xp_3.avif', true, 'level', 10, gen_random_uuid(), NOW(), NOW()),
('Adventurer', 'Reached user level 15', true, 'badges/placeholders/xp_4.avif', true, 'level', 15, gen_random_uuid(), NOW(), NOW()),
('Regular', 'Reached user level 20', true, 'badges/placeholders/xp_5.avif', true, 'level', 20, gen_random_uuid(), NOW(), NOW()),
('Veteran', 'Reached user level 30', true, 'badges/placeholders/xp_6.avif', true, 'level', 30, gen_random_uuid(), NOW(), NOW()),
('Expert', 'Reached user level 40', true, 'badges/placeholders/xp_7.avif', true, 'level', 40, gen_random_uuid(), NOW(), NOW()),
('Master', 'Reached user level 50', true, 'badges/placeholders/xp_8.avif', true, 'level', 50, gen_random_uuid(), NOW(), NOW()),
('Elite', 'Reached user level 60', true, 'badges/placeholders/xp_9.avif', true, 'level', 60, gen_random_uuid(), NOW(), NOW()),
('Champion', 'Reached user level 70', true, 'badges/placeholders/xp_10.avif', true, 'level', 70, gen_random_uuid(), NOW(), NOW()),
('Hero', 'Reached user level 80', true, 'badges/placeholders/xp_11.avif', true, 'level', 80, gen_random_uuid(), NOW(), NOW()),
('Legend', 'Reached user level 85', true, 'badges/placeholders/xp_12.avif', true, 'level', 85, gen_random_uuid(), NOW(), NOW()),
('Mythic', 'Reached user level 90', true, 'badges/placeholders/xp_13.avif', true, 'level', 90, gen_random_uuid(), NOW(), NOW()),
('Immortal', 'Reached user level 95', true, 'badges/placeholders/xp_14.avif', true, 'level', 95, gen_random_uuid(), NOW(), NOW()),
('Ascended', 'Reached user level 98', true, 'badges/placeholders/xp_15.avif', true, 'level', 98, gen_random_uuid(), NOW(), NOW()),
('Celestial', 'Reached the ultimate level 100', true, 'badges/placeholders/xp_16.avif', true, 'level', 100, gen_random_uuid(), NOW(), NOW());

-- 3. Rating Milestones (12 Badges)
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('Casual Listener', 'Rated 5 themes', true, 'badges/placeholders/rating_1.avif', true, 'ratings', 5, gen_random_uuid(), NOW(), NOW()),
('Active Auditor', 'Rated 10 themes', true, 'badges/placeholders/rating_2.avif', true, 'ratings', 10, gen_random_uuid(), NOW(), NOW()),
('Music Enthusiast', 'Rated 25 themes', true, 'badges/placeholders/rating_3.avif', true, 'ratings', 25, gen_random_uuid(), NOW(), NOW()),
('Theme Seeker', 'Rated 50 themes', true, 'badges/placeholders/rating_4.avif', true, 'ratings', 50, gen_random_uuid(), NOW(), NOW()),
('Melody Hunter', 'Rated 100 themes', true, 'badges/placeholders/rating_5.avif', true, 'ratings', 100, gen_random_uuid(), NOW(), NOW()),
('Soundtrack Collector', 'Rated 250 themes', true, 'badges/placeholders/rating_6.avif', true, 'ratings', 250, gen_random_uuid(), NOW(), NOW()),
('Rhythm Scholar', 'Rated 500 themes', true, 'badges/placeholders/rating_7.avif', true, 'ratings', 500, gen_random_uuid(), NOW(), NOW()),
('Discography Explorer', 'Rated 1,000 themes', true, 'badges/placeholders/rating_8.avif', true, 'ratings', 1000, gen_random_uuid(), NOW(), NOW()),
('Archive Guardian', 'Rated 2,500 themes', true, 'badges/placeholders/rating_9.avif', true, 'ratings', 2500, gen_random_uuid(), NOW(), NOW()),
('Symphony Specialist', 'Rated 5,000 themes', true, 'badges/placeholders/rating_10.avif', true, 'ratings', 5000, gen_random_uuid(), NOW(), NOW()),
('Musical Virtuoso', 'Rated 7,500 themes', true, 'badges/placeholders/rating_11.avif', true, 'ratings', 7500, gen_random_uuid(), NOW(), NOW()),
('Grand Maestro', 'Rated 10,000 themes', true, 'badges/placeholders/rating_12.avif', true, 'ratings', 10000, gen_random_uuid(), NOW(), NOW());

-- 4. Commenting Milestones (12 Badges)
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('First Contact', 'Posted your first comment', true, 'badges/placeholders/comment_1.avif', true, 'comments', 1, gen_random_uuid(), NOW(), NOW()),
('Vocal Member', 'Posted 5 comments', true, 'badges/placeholders/comment_2.avif', true, 'comments', 5, gen_random_uuid(), NOW(), NOW()),
('Active Talker', 'Posted 10 comments', true, 'badges/placeholders/comment_3.avif', true, 'comments', 10, gen_random_uuid(), NOW(), NOW()),
('Frequent Debater', 'Posted 25 comments', true, 'badges/placeholders/comment_4.avif', true, 'comments', 25, gen_random_uuid(), NOW(), NOW()),
('Chatterbox', 'Posted 50 comments', true, 'badges/placeholders/comment_5.avif', true, 'comments', 50, gen_random_uuid(), NOW(), NOW()),
('Community Participant', 'Posted 100 comments', true, 'badges/placeholders/comment_6.avif', true, 'comments', 100, gen_random_uuid(), NOW(), NOW()),
('Opinionated Critic', 'Posted 250 comments', true, 'badges/placeholders/comment_7.avif', true, 'comments', 250, gen_random_uuid(), NOW(), NOW()),
('Community Pillar', 'Posted 500 comments', true, 'badges/placeholders/comment_8.avif', true, 'comments', 500, gen_random_uuid(), NOW(), NOW()),
('Wise Mentor', 'Posted 1,000 comments', true, 'badges/placeholders/comment_9.avif', true, 'comments', 1000, gen_random_uuid(), NOW(), NOW()),
('Renowned Speaker', 'Posted 2,000 comments', true, 'badges/placeholders/comment_10.avif', true, 'comments', 2000, gen_random_uuid(), NOW(), NOW()),
('Legendary Orator', 'Posted 3,500 comments', true, 'badges/placeholders/comment_11.avif', true, 'comments', 3500, gen_random_uuid(), NOW(), NOW()),
('Universal Oracle', 'Posted 5,000 comments', true, 'badges/placeholders/comment_12.avif', true, 'comments', 5000, gen_random_uuid(), NOW(), NOW());

-- 5. Special Badges
INSERT INTO badges (name, description, is_active, icon, is_automatic, requirement_type, requirement_value, uuid, created_at, updated_at) VALUES
('AniFan', 'Connected your AniList account', true, 'badges/placeholders/anifan.avif', true, 'anilist', 1, gen_random_uuid(), NOW(), NOW());
