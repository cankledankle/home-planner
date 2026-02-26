-- Add partial unique index for website files with slots
CREATE UNIQUE INDEX IF NOT EXISTS idx_files_plan_slot_website 
ON files (plan_id, slot) 
WHERE category = 'website' AND slot IS NOT NULL;
