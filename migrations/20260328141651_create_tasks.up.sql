CREATE TYPE task_status AS ENUM ('new', 'inProgress', 'completed');

CREATE TABLE IF NOT EXISTS tasks(
    tid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status task_status NOT NULL DEFAULT 'new',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_tasks_user
        FOREIGN KEY (user_id) 
        REFERENCES users(uid)
        ON DELETE CASCADE
);