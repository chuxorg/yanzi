ALTER TABLE intents ADD COLUMN IF NOT EXISTS class TEXT;
ALTER TABLE intents ADD COLUMN IF NOT EXISTS type TEXT;
ALTER TABLE intents ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE intents ADD COLUMN IF NOT EXISTS metadata TEXT;

UPDATE intents
SET
	class = 'intent',
	type = 'prompt',
	title = CASE
		WHEN title IS NULL OR TRIM(title) = '' THEN 'legacy artifact'
		ELSE title
	END,
	content = CASE
		WHEN content IS NULL OR content = '' THEN prompt
		ELSE content
	END,
	metadata = CASE
		WHEN metadata IS NULL THEN meta
		ELSE metadata
	END
WHERE class IS NULL OR type IS NULL OR content IS NULL OR metadata IS NULL;
