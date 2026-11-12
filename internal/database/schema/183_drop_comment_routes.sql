-- phase: post

-- Automatic comment delivery used to be stored per scope in comment_routes.
-- It is now configured as comments.agents in gavel settings (gavel_settings
-- layers and .gavel.yaml files), so the HCL no longer declares the table. Atlas
-- leaves undeclared tables in place (drops are suppressed for the shared
-- database), so this one-time script removes it from an existing database.
DROP TABLE IF EXISTS public.comment_routes;
