-- One-time migration after deploying the runner with restored manual pages.
-- Prefer bin/refresh-man-cache to preview and confirm within a transaction.
-- Match the entire notice, including its final newline; retain successful
-- solutions, real manual output, and all other cached results.
DELETE FROM challenges
WHERE correct = 0
  AND output = 'This system has been minimized by removing packages and content that are
not required on a system that users do not log into.

To restore this content, including manpages, you can run the ''unminimize''
command. You will still need to ensure the ''man-db'' package is installed.
';
