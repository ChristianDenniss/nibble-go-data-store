-- Development-only bootstrap credential for the seeded root account.
-- Replace/remove this before deploying anywhere public.
UPDATE accounts
SET password_hash = '$2a$10$2fMLYbvQ0EjI0fHZuv5z.u173zO3CjiY4ln5SHGkcuGObdwXjQY9u',
    role = 'root',
    updated_at = now()
WHERE lower(email) = lower('aottgpvp@gmail.com');
