-- Idempotent local/admin bootstrap. If the account already exists (including
-- an OAuth-created account), only its role is elevated.
DO $$
DECLARE existing_id TEXT;
BEGIN
    SELECT id INTO existing_id FROM accounts WHERE lower(email) = lower('aottgpvp@gmail.com') LIMIT 1;
    IF existing_id IS NULL THEN
        INSERT INTO accounts (id, name, email, phone, role, password_hash)
        VALUES ('acct_aottgpvp_root', 'Aottg', 'aottgpvp@gmail.com', '', 'root', '$2a$10$2fMLYbvQ0EjI0fHZuv5z.u173zO3CjiY4ln5SHGkcuGObdwXjQY9u');
    ELSE
        UPDATE accounts SET role = 'root', password_hash = '$2a$10$2fMLYbvQ0EjI0fHZuv5z.u173zO3CjiY4ln5SHGkcuGObdwXjQY9u', updated_at = now() WHERE id = existing_id;
    END IF;
END $$;
