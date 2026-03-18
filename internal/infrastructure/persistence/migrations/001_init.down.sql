-- Rollback 001_init

DROP INDEX IF EXISTS idx_transactions_user_id_created_at;
DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS idx_skills_category;
DROP INDEX IF EXISTS idx_skills_is_installed;
DROP INDEX IF EXISTS idx_skills_is_recommended;
DROP TABLE IF EXISTS skills;

DROP INDEX IF EXISTS idx_resources_project_id;
DROP INDEX IF EXISTS idx_resources_type;
DROP TABLE IF EXISTS resources;

DROP INDEX IF EXISTS idx_messages_project_id_created_at;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS idx_projects_user_id;
DROP INDEX IF EXISTS idx_projects_status;
DROP INDEX IF EXISTS idx_projects_updated_at;
DROP TABLE IF EXISTS projects;

DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
