CREATE TABLE teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    is_active BOOLEAN NOT NULL
);

CREATE TABLE team_members (
    team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_name, user_id)
);

CREATE TABLE pull_requests (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    assigned_reviewers TEXT[],
    created_at TIMESTAMP NOT NULL,
    merged_at TIMESTAMP
);

-- Индексы
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_team_members_team ON team_members(team_name);
CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_reviewers ON pull_requests USING GIN (assigned_reviewers);