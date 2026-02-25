-- 流程範本
CREATE TABLE workflow_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(50) NOT NULL DEFAULT 'primary',
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_workflow_templates_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_workflow_templates_user_id ON workflow_templates(user_id);

-- 流程階段
CREATE TABLE workflow_phases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_template_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    duration_days INT NOT NULL DEFAULT 1,
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_workflow_phases_template FOREIGN KEY (workflow_template_id) REFERENCES workflow_templates(id) ON DELETE CASCADE
);
CREATE INDEX idx_workflow_phases_template_id ON workflow_phases(workflow_template_id);

-- 合作項目
CREATE TABLE collaboration_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    price NUMERIC(12,2) NOT NULL DEFAULT 0,
    parent_id UUID,
    workflow_id UUID,
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_collaboration_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_collaboration_items_parent FOREIGN KEY (parent_id) REFERENCES collaboration_items(id) ON DELETE SET NULL,
    CONSTRAINT fk_collaboration_items_workflow FOREIGN KEY (workflow_id) REFERENCES workflow_templates(id) ON DELETE SET NULL
);
CREATE INDEX idx_collaboration_items_user_id ON collaboration_items(user_id);
CREATE INDEX idx_collaboration_items_parent_id ON collaboration_items(parent_id);

-- 案件階段實例
CREATE TABLE case_phases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    start_date DATE,
    end_date DATE,
    duration_days INT NOT NULL DEFAULT 1,
    "order" INT NOT NULL DEFAULT 0,
    workflow_phase_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_case_phases_case FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    CONSTRAINT fk_case_phases_workflow_phase FOREIGN KEY (workflow_phase_id) REFERENCES workflow_phases(id) ON DELETE SET NULL
);
CREATE INDEX idx_case_phases_case_id ON case_phases(case_id);
