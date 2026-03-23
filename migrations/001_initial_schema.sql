-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- User Management
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}'
);

-- Roles and Permissions
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE TABLE team_members (
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

-- Runbooks
CREATE TABLE runbooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    definition JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    team_id UUID REFERENCES teams(id),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    CONSTRAINT unique_runbook_name_per_team UNIQUE (name, team_id)
);

-- Runbook Versions (for version control)
CREATE TABLE runbook_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    runbook_id UUID REFERENCES runbooks(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    definition JSONB NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    change_description TEXT,
    is_current BOOLEAN DEFAULT false,
    CONSTRAINT unique_runbook_version UNIQUE (runbook_id, version)
);

-- Triggers
CREATE TABLE triggers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    runbook_id UUID REFERENCES runbooks(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'webhook', 'schedule', 'alert', 'manual'
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id)
);

-- Executions
CREATE TABLE executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    runbook_id UUID REFERENCES runbooks(id),
    runbook_version INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    trigger_type VARCHAR(50) NOT NULL,
    trigger_info JSONB NOT NULL DEFAULT '{}',
    context JSONB NOT NULL DEFAULT '{}',
    started_by UUID REFERENCES users(id),
    workflow_id VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    error_message TEXT,
    error_details JSONB,
    metadata JSONB DEFAULT '{}'
);

-- Execution Steps
CREATE TABLE execution_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    execution_id UUID REFERENCES executions(id) ON DELETE CASCADE,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    input JSONB,
    output JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    order_index INTEGER NOT NULL,
    CONSTRAINT unique_execution_step UNIQUE (execution_id, node_id)
);

-- Integrations
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL, -- 'kubernetes', 'datadog', 'slack', 'github', etc.
    config JSONB NOT NULL,
    secret_ref VARCHAR(255), -- Reference to secret in Vault
    team_id UUID REFERENCES teams(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id),
    CONSTRAINT unique_integration_name_per_team UNIQUE (name, team_id)
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

-- Templates
CREATE TABLE runbook_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    definition JSONB NOT NULL,
    is_public BOOLEAN DEFAULT false,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    usage_count INTEGER DEFAULT 0
);

-- Schedules
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trigger_id UUID REFERENCES triggers(id) ON DELETE CASCADE,
    cron_expression VARCHAR(100) NOT NULL,
    timezone VARCHAR(50) DEFAULT 'UTC',
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_runbooks_team_id ON runbooks(team_id);
CREATE INDEX idx_runbooks_created_by ON runbooks(created_by);
CREATE INDEX idx_runbooks_tags ON runbooks USING GIN(tags);
CREATE INDEX idx_runbooks_active ON runbooks(is_active);
CREATE INDEX idx_executions_runbook_id ON executions(runbook_id);
CREATE INDEX idx_executions_status ON executions(status);
CREATE INDEX idx_executions_started_at ON executions(started_at);
CREATE INDEX idx_executions_started_by ON executions(started_by);
CREATE INDEX idx_execution_steps_execution_id ON execution_steps(execution_id);
CREATE INDEX idx_execution_steps_status ON execution_steps(status);
CREATE INDEX idx_triggers_runbook_id ON triggers(runbook_id);
CREATE INDEX idx_triggers_type ON triggers(type);
CREATE INDEX idx_triggers_active ON triggers(is_active);
CREATE INDEX idx_integrations_team_id ON integrations(team_id);
CREATE INDEX idx_integrations_type ON integrations(type);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);

-- JSONB indexes for complex queries
CREATE INDEX idx_runbooks_definition ON runbooks USING GIN(definition);
CREATE INDEX idx_executions_context ON executions USING GIN(context);
CREATE INDEX idx_executions_trigger_info ON executions USING GIN(trigger_info);
CREATE INDEX idx_execution_steps_input ON execution_steps USING GIN(input);
CREATE INDEX idx_execution_steps_output ON execution_steps USING GIN(output);

-- Full-text search indexes
CREATE INDEX idx_runbooks_name_search ON runbooks USING GIN(to_tsvector('english', name));
CREATE INDEX idx_runbooks_description_search ON runbooks USING GIN(to_tsvector('english', description));

-- Update timestamp triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_runbooks_updated_at BEFORE UPDATE ON runbooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_triggers_updated_at BEFORE UPDATE ON triggers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_integrations_updated_at BEFORE UPDATE ON integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_runbook_templates_updated_at BEFORE UPDATE ON runbook_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles
INSERT INTO roles (name, description, permissions) VALUES
('admin', 'System administrator', '["*"]'),
('owner', 'Team owner', '["runbooks:*", "team:*", "executions:*"]'),
('editor', 'Runbook editor', '["runbooks:read", "runbooks:write", "executions:*"]'),
('viewer', 'Read-only access', '["runbooks:read", "executions:read"]');

-- Insert default admin user (change password in production)
INSERT INTO users (email, name, password_hash) VALUES
('admin@runbook-engine.com', 'System Administrator', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjdLrrLjYz8d5Y5N5Y5Y5Y5Y5Y5Y5Y');

-- Assign admin role to admin user
INSERT INTO user_roles (user_id, role_id, assigned_by) 
SELECT u.id, r.id, u.id FROM users u, roles r 
WHERE u.email = 'admin@runbook-engine.com' AND r.name = 'admin';

-- Insert default team
INSERT INTO teams (name, description) VALUES
('Default Team', 'Default team for new users');

-- Add admin user to default team as owner
INSERT INTO team_members (team_id, user_id, role) 
SELECT t.id, u.id, 'owner' FROM teams t, users u 
WHERE t.name = 'Default Team' AND u.email = 'admin@runbook-engine.com';
