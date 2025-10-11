SELECT pg_catalog.set_config('search_path', '', false);
CREATE TABLE public.agent_ai_interactions (
    id uuid NOT NULL,
    group_id uuid NOT NULL,
    session_id uuid NOT NULL,
    contact_id uuid NOT NULL,
    channel_id uuid NOT NULL,
    tenant_id character varying(255) NOT NULL,
    concatenated_content text NOT NULL,
    message_count bigint NOT NULL,
    enrichment_count bigint NOT NULL,
    sent_to_ai boolean DEFAULT false NOT NULL,
    ai_response text,
    a_iprovider character varying(50),
    ai_model character varying(100),
    processing_time_ms bigint,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    sent_at timestamp with time zone,
    response_received_at timestamp with time zone
);
CREATE TABLE public.agent_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    agent_id uuid NOT NULL,
    session_id uuid NOT NULL,
    role_in_session text,
    joined_at timestamp with time zone NOT NULL,
    left_at timestamp with time zone,
    is_active boolean DEFAULT true,
    metadata jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.agents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    user_id uuid,
    tenant_id text NOT NULL,
    name text NOT NULL,
    email text,
    type text DEFAULT 'human'::text NOT NULL,
    status text DEFAULT 'offline'::text NOT NULL,
    active boolean DEFAULT true,
    config jsonb,
    sessions_handled bigint DEFAULT 0,
    average_response_ms bigint DEFAULT 0,
    last_activity_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.automations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    automation_type character varying(100) DEFAULT 'pipeline_automation'::character varying NOT NULL,
    pipeline_id uuid,
    tenant_id character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    trigger character varying(100) NOT NULL,
    conditions jsonb DEFAULT '[]'::jsonb,
    actions jsonb DEFAULT '[]'::jsonb,
    priority bigint DEFAULT 0 NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    schedule jsonb,
    last_executed timestamp with time zone,
    next_execution timestamp with time zone
);
CREATE TABLE public.billing_accounts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    payment_status text DEFAULT 'pending'::text NOT NULL,
    payment_methods jsonb,
    billing_email text NOT NULL,
    suspended boolean DEFAULT false,
    suspended_at timestamp with time zone,
    suspension_reason text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.channel_types (
    id bigint NOT NULL,
    name text NOT NULL,
    description text,
    provider text NOT NULL,
    configuration jsonb,
    active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.channels (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'inactive'::text,
    external_id text,
    config jsonb,
    webhook_id text,
    webhook_url text,
    webhook_configured_at timestamp with time zone,
    webhook_active boolean DEFAULT false,
    pipeline_id uuid,
    session_timeout_minutes bigint,
    ai_enabled boolean DEFAULT false,
    ai_agents_enabled boolean DEFAULT false,
    allow_groups boolean DEFAULT false,
    tracking_enabled boolean DEFAULT false,
    debounce_timeout_ms bigint DEFAULT 15000 NOT NULL,
    messages_received bigint DEFAULT 0,
    messages_sent bigint DEFAULT 0,
    last_message_at timestamp with time zone,
    last_error_at timestamp with time zone,
    last_error text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.chats (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    chat_type text NOT NULL,
    external_id text,
    subject character varying(255),
    description text,
    participants jsonb NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    last_message_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.contact_event_store (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_type character varying(50) DEFAULT 'contact'::character varying NOT NULL,
    event_type character varying(100) NOT NULL,
    event_version character varying(10) DEFAULT 'v1'::character varying NOT NULL,
    sequence_number bigint NOT NULL,
    event_data jsonb NOT NULL,
    metadata jsonb,
    occurred_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id character varying(255) NOT NULL,
    project_id uuid,
    causation_id uuid,
    correlation_id uuid
);
CREATE TABLE public.contact_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    contact_id uuid NOT NULL,
    session_id uuid,
    tenant_id text NOT NULL,
    event_type text NOT NULL,
    category text NOT NULL,
    priority text NOT NULL,
    title text,
    description text,
    payload jsonb,
    metadata jsonb,
    source text NOT NULL,
    triggered_by uuid,
    integration_source text,
    is_realtime boolean DEFAULT true,
    delivered boolean DEFAULT false,
    delivered_at timestamp with time zone,
    read boolean DEFAULT false,
    read_at timestamp with time zone,
    visible_to_client boolean DEFAULT true,
    visible_to_agent boolean DEFAULT true,
    expires_at timestamp with time zone,
    occurred_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.contact_lists (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    description text,
    logical_operator text DEFAULT 'AND'::text NOT NULL,
    is_static boolean DEFAULT false NOT NULL,
    contact_count bigint DEFAULT 0 NOT NULL,
    last_calculated_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.contact_pipeline_statuses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    contact_id uuid NOT NULL,
    pipeline_id uuid NOT NULL,
    status_id uuid NOT NULL,
    tenant_id text NOT NULL,
    entered_at timestamp with time zone NOT NULL,
    exited_at timestamp with time zone,
    duration bigint,
    notes text,
    metadata jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.contact_snapshots (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    aggregate_id uuid NOT NULL,
    snapshot_data jsonb NOT NULL,
    last_sequence_number bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    tenant_id character varying(255) NOT NULL
);
CREATE TABLE public.contacts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text,
    email text,
    phone text,
    external_id text,
    source_channel text,
    language text DEFAULT 'en'::text,
    timezone text,
    tags jsonb,
    profile_picture_url text,
    profile_picture_fetched_at timestamp with time zone,
    first_interaction_at timestamp with time zone,
    last_interaction_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.credentials (
    id uuid NOT NULL,
    tenant_id character varying(255) NOT NULL,
    project_id uuid,
    credential_type character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    encrypted_value_ciphertext text NOT NULL,
    encrypted_value_nonce text NOT NULL,
    oauth_access_token_ciphertext text,
    oauth_access_token_nonce text,
    oauth_refresh_token_ciphertext text,
    oauth_refresh_token_nonce text,
    oauth_token_type character varying(20),
    oauth_expires_at timestamp with time zone,
    metadata jsonb DEFAULT '{}'::jsonb,
    is_active boolean DEFAULT true NOT NULL,
    expires_at timestamp with time zone,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.domain_event_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_type text NOT NULL,
    tenant_id text NOT NULL,
    project_id uuid,
    user_id uuid,
    payload jsonb,
    occurred_at timestamp with time zone NOT NULL,
    published_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.message_enrichments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    message_group_id uuid NOT NULL,
    content_type character varying(50) NOT NULL,
    provider character varying(50) NOT NULL,
    media_url text NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    extracted_text text,
    metadata jsonb,
    processing_time_ms bigint,
    error text,
    context character varying(50),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone
);
CREATE TABLE public.message_groups (
    id uuid NOT NULL,
    contact_id uuid NOT NULL,
    channel_id uuid NOT NULL,
    session_id uuid NOT NULL,
    tenant_id character varying(255) NOT NULL,
    message_ids text[] NOT NULL,
    status character varying(50) NOT NULL,
    started_at timestamp with time zone NOT NULL,
    completed_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);
CREATE TABLE public.messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id text NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    channel_type_id bigint,
    from_me boolean DEFAULT false,
    channel_id uuid NOT NULL,
    chat_id uuid,
    contact_id uuid NOT NULL,
    session_id uuid,
    content_type text DEFAULT 'text'::text NOT NULL,
    text text,
    media_url text,
    media_mimetype text,
    channel_message_id text,
    reply_to_id uuid,
    status text DEFAULT 'sent'::text,
    language text,
    agent_id uuid,
    metadata jsonb,
    mentions text[],
    delivered_at timestamp with time zone,
    read_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.notes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    contact_id uuid NOT NULL,
    session_id uuid,
    tenant_id text NOT NULL,
    author_id uuid NOT NULL,
    author_type text NOT NULL,
    author_name text NOT NULL,
    content text NOT NULL,
    note_type text NOT NULL,
    priority text DEFAULT 'normal'::text NOT NULL,
    visible_to_client boolean DEFAULT false,
    pinned boolean DEFAULT false,
    tags text[],
    mentions jsonb,
    attachments text[],
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.outbox_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id uuid NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_type character varying(100) NOT NULL,
    event_type character varying(100) NOT NULL,
    event_version character varying(20) DEFAULT 'v1'::character varying NOT NULL,
    event_data jsonb NOT NULL,
    metadata jsonb,
    tenant_id character varying(100),
    project_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    retry_count bigint DEFAULT 0 NOT NULL,
    last_error text,
    last_retry_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.pipeline_statuses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    pipeline_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    color text,
    status_type text NOT NULL,
    "position" bigint DEFAULT 0,
    active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.pipelines (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    description text,
    color text,
    "position" bigint DEFAULT 0,
    active boolean DEFAULT true,
    session_timeout_minutes bigint,
    enable_ai_summary boolean DEFAULT false,
    a_iprovider text,
    ai_model text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.processed_events (
    id bigint NOT NULL,
    event_id uuid NOT NULL,
    consumer_name character varying(100) NOT NULL,
    processed_at timestamp with time zone DEFAULT now() NOT NULL,
    processing_duration_ms bigint
);
CREATE SEQUENCE public.processed_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.processed_events_id_seq OWNED BY public.processed_events.id;
CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    billing_account_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    description text,
    configuration jsonb,
    active boolean DEFAULT true,
    session_timeout_minutes bigint DEFAULT 30 NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    contact_id uuid NOT NULL,
    tenant_id text NOT NULL,
    channel_type_id bigint,
    pipeline_id uuid,
    started_at timestamp with time zone NOT NULL,
    ended_at timestamp with time zone,
    status text DEFAULT 'active'::text,
    end_reason text,
    timeout_duration bigint DEFAULT '1800000000000'::bigint,
    last_activity_at timestamp with time zone NOT NULL,
    message_count bigint DEFAULT 0,
    messages_from_contact bigint DEFAULT 0,
    messages_from_agent bigint DEFAULT 0,
    duration_seconds bigint DEFAULT 0,
    first_contact_message_at timestamp with time zone,
    first_agent_response_at timestamp with time zone,
    agent_response_time_seconds bigint,
    contact_wait_time_seconds bigint,
    agent_ids jsonb,
    agent_transfers bigint DEFAULT 0,
    summary text,
    sentiment text,
    sentiment_score numeric,
    topics jsonb,
    next_steps jsonb,
    key_entities jsonb,
    resolved boolean DEFAULT false,
    escalated boolean DEFAULT false,
    converted boolean DEFAULT false,
    outcome_tags jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.tracking_enrichments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tracking_id uuid NOT NULL,
    tenant_id text NOT NULL,
    source text NOT NULL,
    ad_account_id text,
    ad_account_name text,
    campaign_id text,
    campaign_name text,
    adset_id text,
    adset_name text,
    ad_id text,
    ad_name text,
    ad_creative_id text,
    creative_type text,
    creative_format text,
    creative_body text,
    creative_title text,
    creative_url text,
    targeting_data jsonb,
    audience_name text,
    impressions bigint,
    clicks bigint,
    spend numeric,
    ctr numeric,
    cpc numeric,
    raw_api_data jsonb,
    enriched_at timestamp with time zone NOT NULL,
    enrichment_type text,
    api_version text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);
CREATE TABLE public.trackings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    contact_id uuid NOT NULL,
    session_id uuid,
    tenant_id text NOT NULL,
    project_id uuid NOT NULL,
    source text NOT NULL,
    platform text NOT NULL,
    campaign text,
    ad_id text,
    ad_url text,
    click_id text,
    conversion_data text,
    utm_source text,
    utm_medium text,
    utm_campaign text,
    utm_term text,
    utm_content text,
    metadata jsonb,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);
CREATE TABLE public.user_api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    key_hash text NOT NULL,
    active boolean DEFAULT true,
    last_used timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    status text DEFAULT 'active'::text,
    role text DEFAULT 'user'::text,
    settings jsonb,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_users_role CHECK ((role = ANY (ARRAY['admin'::text, 'user'::text, 'manager'::text, 'readonly'::text])))
);
CREATE TABLE public.webhook_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    tenant_id text NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    events text[],
    subscribe_contact_events boolean DEFAULT false,
    contact_event_types text[],
    contact_event_categories text[],
    active boolean DEFAULT true,
    secret text,
    headers jsonb,
    retry_count bigint DEFAULT 3,
    timeout_seconds bigint DEFAULT 30,
    last_triggered_at timestamp with time zone,
    last_success_at timestamp with time zone,
    last_failure_at timestamp with time zone,
    success_count bigint DEFAULT 0,
    failure_count bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
ALTER TABLE ONLY public.processed_events ALTER COLUMN id SET DEFAULT nextval('public.processed_events_id_seq'::regclass);
ALTER TABLE ONLY public.agent_ai_interactions
    ADD CONSTRAINT agent_ai_interactions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.agent_sessions
    ADD CONSTRAINT agent_sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.automations
    ADD CONSTRAINT automations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.billing_accounts
    ADD CONSTRAINT billing_accounts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.channel_types
    ADD CONSTRAINT channel_types_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.channels
    ADD CONSTRAINT channels_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.chats
    ADD CONSTRAINT chats_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contact_event_store
    ADD CONSTRAINT contact_event_store_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contact_events
    ADD CONSTRAINT contact_events_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contact_lists
    ADD CONSTRAINT contact_lists_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contact_pipeline_statuses
    ADD CONSTRAINT contact_pipeline_statuses_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contact_snapshots
    ADD CONSTRAINT contact_snapshots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.contacts
    ADD CONSTRAINT contacts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.credentials
    ADD CONSTRAINT credentials_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.domain_event_logs
    ADD CONSTRAINT domain_event_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.message_enrichments
    ADD CONSTRAINT message_enrichments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.message_groups
    ADD CONSTRAINT message_groups_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.pipeline_statuses
    ADD CONSTRAINT pipeline_statuses_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.pipelines
    ADD CONSTRAINT pipelines_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.processed_events
    ADD CONSTRAINT processed_events_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.tracking_enrichments
    ADD CONSTRAINT tracking_enrichments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.trackings
    ADD CONSTRAINT trackings_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.user_api_keys
    ADD CONSTRAINT user_api_keys_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.webhook_subscriptions
    ADD CONSTRAINT webhook_subscriptions_pkey PRIMARY KEY (id);
CREATE INDEX idx_agent_sessions_agent_id ON public.agent_sessions USING btree (agent_id);
CREATE INDEX idx_agent_sessions_deleted_at ON public.agent_sessions USING btree (deleted_at);
CREATE INDEX idx_agent_sessions_is_active ON public.agent_sessions USING btree (is_active);
CREATE INDEX idx_agent_sessions_joined_at ON public.agent_sessions USING btree (joined_at);
CREATE INDEX idx_agent_sessions_left_at ON public.agent_sessions USING btree (left_at);
CREATE INDEX idx_agent_sessions_session_id ON public.agent_sessions USING btree (session_id);
CREATE INDEX idx_agents_active ON public.agents USING btree (active);
CREATE INDEX idx_agents_config ON public.agents USING gin (config);
CREATE INDEX idx_agents_created ON public.agents USING btree (created_at);
CREATE INDEX idx_agents_deleted ON public.agents USING btree (deleted_at);
CREATE INDEX idx_agents_email ON public.agents USING btree (email);
CREATE INDEX idx_agents_last_activity ON public.agents USING btree (last_activity_at);
CREATE INDEX idx_agents_name ON public.agents USING btree (name);
CREATE INDEX idx_agents_project ON public.agents USING btree (project_id);
CREATE INDEX idx_agents_status ON public.agents USING btree (status);
CREATE INDEX idx_agents_tenant ON public.agents USING btree (tenant_id);
CREATE INDEX idx_agents_tenant_active ON public.agents USING btree (tenant_id, active);
CREATE INDEX idx_agents_tenant_status ON public.agents USING btree (tenant_id, status);
CREATE INDEX idx_agents_tenant_type ON public.agents USING btree (tenant_id, type);
CREATE INDEX idx_agents_type ON public.agents USING btree (type);
CREATE INDEX idx_agents_updated ON public.agents USING btree (updated_at);
CREATE INDEX idx_agents_user ON public.agents USING btree (user_id);
CREATE INDEX idx_ai_agent_history_contact ON public.agent_ai_interactions USING btree (contact_id);
CREATE INDEX idx_ai_agent_history_group_id ON public.agent_ai_interactions USING btree (group_id);
CREATE INDEX idx_ai_agent_history_sent_to_ai ON public.agent_ai_interactions USING btree (sent_to_ai);
CREATE INDEX idx_ai_agent_history_session_id ON public.agent_ai_interactions USING btree (session_id);
CREATE INDEX idx_ai_agent_history_tenant ON public.agent_ai_interactions USING btree (tenant_id);
CREATE INDEX idx_automations_enabled ON public.automations USING btree (enabled);
CREATE INDEX idx_automations_last_executed ON public.automations USING btree (last_executed);
CREATE INDEX idx_automations_next_execution ON public.automations USING btree (next_execution);
CREATE INDEX idx_automations_pipeline ON public.automations USING btree (pipeline_id);
CREATE INDEX idx_automations_priority ON public.automations USING btree (priority);
CREATE INDEX idx_automations_tenant ON public.automations USING btree (tenant_id);
CREATE INDEX idx_automations_trigger ON public.automations USING btree (trigger);
CREATE INDEX idx_automations_type ON public.automations USING btree (automation_type);
CREATE INDEX idx_billing_accounts_deleted_at ON public.billing_accounts USING btree (deleted_at);
CREATE INDEX idx_billing_accounts_payment_status ON public.billing_accounts USING btree (payment_status);
CREATE INDEX idx_billing_accounts_suspended ON public.billing_accounts USING btree (suspended);
CREATE INDEX idx_billing_accounts_user_id ON public.billing_accounts USING btree (user_id);
CREATE INDEX idx_channel_types_active ON public.channel_types USING btree (active);
CREATE INDEX idx_channel_types_deleted_at ON public.channel_types USING btree (deleted_at);
CREATE UNIQUE INDEX idx_channel_types_name ON public.channel_types USING btree (name);
CREATE INDEX idx_channels_ai_agents ON public.channels USING btree (ai_agents_enabled);
CREATE INDEX idx_channels_ai_enabled ON public.channels USING btree (ai_enabled);
CREATE INDEX idx_channels_allow_groups ON public.channels USING btree (allow_groups);
CREATE INDEX idx_channels_config ON public.channels USING gin (config);
CREATE INDEX idx_channels_created ON public.channels USING btree (created_at);
CREATE INDEX idx_channels_deleted ON public.channels USING btree (deleted_at);
CREATE INDEX idx_channels_external_id ON public.channels USING btree (external_id);
CREATE INDEX idx_channels_last_error ON public.channels USING btree (last_error_at);
CREATE INDEX idx_channels_last_message ON public.channels USING btree (last_message_at);
CREATE INDEX idx_channels_name ON public.channels USING btree (name);
CREATE INDEX idx_channels_pipeline ON public.channels USING btree (pipeline_id);
CREATE INDEX idx_channels_project ON public.channels USING btree (project_id);
CREATE INDEX idx_channels_project_type ON public.channels USING btree (project_id, type);
CREATE INDEX idx_channels_status ON public.channels USING btree (status);
CREATE INDEX idx_channels_tenant ON public.channels USING btree (tenant_id);
CREATE INDEX idx_channels_tenant_status ON public.channels USING btree (tenant_id, status);
CREATE INDEX idx_channels_tenant_type ON public.channels USING btree (tenant_id, type);
CREATE INDEX idx_channels_timeout ON public.channels USING btree (session_timeout_minutes);
CREATE INDEX idx_channels_tracking_enabled ON public.channels USING btree (tracking_enabled);
CREATE INDEX idx_channels_type ON public.channels USING btree (type);
CREATE INDEX idx_channels_updated ON public.channels USING btree (updated_at);
CREATE INDEX idx_channels_user ON public.channels USING btree (user_id);
CREATE INDEX idx_channels_webhook_active ON public.channels USING btree (webhook_active);
CREATE INDEX idx_channels_webhook_configured ON public.channels USING btree (webhook_configured_at);
CREATE UNIQUE INDEX idx_channels_webhook_id_unique ON public.channels USING btree (webhook_id);
CREATE INDEX idx_channels_webhook_url ON public.channels USING btree (webhook_url);
CREATE INDEX idx_chats_created ON public.chats USING btree (created_at);
CREATE INDEX idx_chats_deleted ON public.chats USING btree (deleted_at);
CREATE INDEX idx_chats_external_id ON public.chats USING btree (external_id);
CREATE INDEX idx_chats_last_message ON public.chats USING btree (last_message_at);
CREATE INDEX idx_chats_participants ON public.chats USING gin (participants);
CREATE INDEX idx_chats_project ON public.chats USING btree (project_id);
CREATE INDEX idx_chats_status ON public.chats USING btree (status);
CREATE INDEX idx_chats_subject ON public.chats USING btree (subject);
CREATE INDEX idx_chats_tenant ON public.chats USING btree (tenant_id);
CREATE INDEX idx_chats_tenant_status ON public.chats USING btree (tenant_id, status);
CREATE INDEX idx_chats_tenant_type ON public.chats USING btree (tenant_id, chat_type);
CREATE INDEX idx_chats_type ON public.chats USING btree (chat_type);
CREATE INDEX idx_chats_updated ON public.chats USING btree (updated_at);
CREATE INDEX idx_contact_events_aggregate ON public.contact_event_store USING btree (aggregate_id, sequence_number);
CREATE INDEX idx_contact_events_category ON public.contact_events USING btree (category);
CREATE INDEX idx_contact_events_contact_id ON public.contact_events USING btree (contact_id);
CREATE INDEX idx_contact_events_correlation ON public.contact_event_store USING btree (correlation_id);
CREATE INDEX idx_contact_events_deleted_at ON public.contact_events USING btree (deleted_at);
CREATE INDEX idx_contact_events_delivered ON public.contact_events USING btree (delivered);
CREATE INDEX idx_contact_events_event_type ON public.contact_events USING btree (event_type);
CREATE INDEX idx_contact_events_expires_at ON public.contact_events USING btree (expires_at);
CREATE INDEX idx_contact_events_is_realtime ON public.contact_events USING btree (is_realtime);
CREATE INDEX idx_contact_events_occurred ON public.contact_event_store USING btree (occurred_at DESC);
CREATE INDEX idx_contact_events_occurred_at ON public.contact_events USING btree (occurred_at);
CREATE INDEX idx_contact_events_priority ON public.contact_events USING btree (priority);
CREATE INDEX idx_contact_events_read ON public.contact_events USING btree (read);
CREATE INDEX idx_contact_events_session_id ON public.contact_events USING btree (session_id);
CREATE INDEX idx_contact_events_source ON public.contact_events USING btree (source);
CREATE INDEX idx_contact_events_tenant ON public.contact_event_store USING btree (tenant_id);
CREATE INDEX idx_contact_events_tenant_id ON public.contact_events USING btree (tenant_id);
CREATE INDEX idx_contact_events_triggered_by ON public.contact_events USING btree (triggered_by);
CREATE INDEX idx_contact_events_type ON public.contact_event_store USING btree (event_type, occurred_at DESC);
CREATE INDEX idx_contact_events_visible_to_agent ON public.contact_events USING btree (visible_to_agent);
CREATE INDEX idx_contact_events_visible_to_client ON public.contact_events USING btree (visible_to_client);
CREATE INDEX idx_contact_lists_deleted_at ON public.contact_lists USING btree (deleted_at);
CREATE INDEX idx_contact_lists_project_id ON public.contact_lists USING btree (project_id);
CREATE INDEX idx_contact_lists_tenant_id ON public.contact_lists USING btree (tenant_id);
CREATE INDEX idx_contact_pipeline_statuses_contact_id ON public.contact_pipeline_statuses USING btree (contact_id);
CREATE INDEX idx_contact_pipeline_statuses_deleted_at ON public.contact_pipeline_statuses USING btree (deleted_at);
CREATE INDEX idx_contact_pipeline_statuses_pipeline_id ON public.contact_pipeline_statuses USING btree (pipeline_id);
CREATE INDEX idx_contact_pipeline_statuses_status_id ON public.contact_pipeline_statuses USING btree (status_id);
CREATE INDEX idx_contact_pipeline_statuses_tenant_id ON public.contact_pipeline_statuses USING btree (tenant_id);
CREATE INDEX idx_contact_snapshots_aggregate ON public.contact_snapshots USING btree (aggregate_id, last_sequence_number DESC);
CREATE INDEX idx_contact_snapshots_tenant ON public.contact_snapshots USING btree (tenant_id, created_at DESC);
CREATE INDEX idx_contacts_created ON public.contacts USING btree (created_at);
CREATE INDEX idx_contacts_deleted ON public.contacts USING btree (deleted_at);
CREATE INDEX idx_contacts_email ON public.contacts USING btree (email);
CREATE INDEX idx_contacts_external_id ON public.contacts USING btree (external_id);
CREATE INDEX idx_contacts_name ON public.contacts USING btree (name);
CREATE INDEX idx_contacts_phone ON public.contacts USING btree (phone);
CREATE INDEX idx_contacts_project_id ON public.contacts USING btree (project_id);
CREATE INDEX idx_contacts_tags ON public.contacts USING gin (tags);
CREATE INDEX idx_contacts_tenant_created ON public.contacts USING btree (tenant_id, created_at);
CREATE INDEX idx_contacts_tenant_deleted ON public.contacts USING btree (tenant_id, deleted_at);
CREATE INDEX idx_contacts_tenant_name ON public.contacts USING btree (tenant_id, name);
CREATE INDEX idx_contacts_updated ON public.contacts USING btree (updated_at);
CREATE INDEX idx_credentials_active ON public.credentials USING btree (is_active);
CREATE INDEX idx_credentials_expires_at ON public.credentials USING btree (expires_at);
CREATE INDEX idx_credentials_project ON public.credentials USING btree (project_id);
CREATE INDEX idx_credentials_tenant ON public.credentials USING btree (tenant_id);
CREATE INDEX idx_credentials_type ON public.credentials USING btree (credential_type);
CREATE INDEX idx_domain_event_logs_aggregate_id ON public.domain_event_logs USING btree (aggregate_id);
CREATE INDEX idx_domain_event_logs_aggregate_type ON public.domain_event_logs USING btree (aggregate_type);
CREATE INDEX idx_domain_event_logs_deleted_at ON public.domain_event_logs USING btree (deleted_at);
CREATE INDEX idx_domain_event_logs_event_type ON public.domain_event_logs USING btree (event_type);
CREATE INDEX idx_domain_event_logs_occurred_at ON public.domain_event_logs USING btree (occurred_at);
CREATE INDEX idx_domain_event_logs_project_id ON public.domain_event_logs USING btree (project_id);
CREATE INDEX idx_domain_event_logs_published_at ON public.domain_event_logs USING btree (published_at);
CREATE INDEX idx_domain_event_logs_tenant_id ON public.domain_event_logs USING btree (tenant_id);
CREATE INDEX idx_domain_event_logs_user_id ON public.domain_event_logs USING btree (user_id);
CREATE INDEX idx_enrichments_content_type ON public.message_enrichments USING btree (content_type);
CREATE INDEX idx_enrichments_created ON public.message_enrichments USING btree (created_at);
CREATE INDEX idx_enrichments_enriched_at ON public.tracking_enrichments USING btree (enriched_at);
CREATE INDEX idx_enrichments_group ON public.message_enrichments USING btree (message_group_id);
CREATE INDEX idx_enrichments_message ON public.message_enrichments USING btree (message_id);
CREATE INDEX idx_enrichments_source ON public.tracking_enrichments USING btree (source);
CREATE INDEX idx_enrichments_status ON public.message_enrichments USING btree (status);
CREATE INDEX idx_enrichments_tenant_id ON public.tracking_enrichments USING btree (tenant_id);
CREATE UNIQUE INDEX idx_enrichments_tracking_id ON public.tracking_enrichments USING btree (tracking_id);
CREATE INDEX idx_message_groups_contact_channel ON public.message_groups USING btree (contact_id, channel_id);
CREATE INDEX idx_message_groups_expires_at ON public.message_groups USING btree (expires_at);
CREATE INDEX idx_message_groups_session ON public.message_groups USING btree (session_id);
CREATE INDEX idx_message_groups_status ON public.message_groups USING btree (status);
CREATE INDEX idx_message_groups_tenant ON public.message_groups USING btree (tenant_id);
CREATE INDEX idx_messages_agent ON public.messages USING btree (agent_id);
CREATE INDEX idx_messages_channel ON public.messages USING btree (channel_id);
CREATE INDEX idx_messages_channel_message_id ON public.messages USING btree (channel_message_id);
CREATE INDEX idx_messages_channel_type ON public.messages USING btree (channel_type_id);
CREATE INDEX idx_messages_chat_id ON public.messages USING btree (chat_id);
CREATE INDEX idx_messages_contact ON public.messages USING btree (contact_id);
CREATE INDEX idx_messages_content_type ON public.messages USING btree (content_type);
CREATE INDEX idx_messages_created ON public.messages USING btree (created_at);
CREATE INDEX idx_messages_deleted ON public.messages USING btree (deleted_at);
CREATE INDEX idx_messages_delivered_at ON public.messages USING btree (delivered_at);
CREATE INDEX idx_messages_from_me ON public.messages USING btree (from_me);
CREATE INDEX idx_messages_metadata ON public.messages USING gin (metadata);
CREATE INDEX idx_messages_project ON public.messages USING btree (project_id);
CREATE INDEX idx_messages_read_at ON public.messages USING btree (read_at);
CREATE INDEX idx_messages_session ON public.messages USING btree (session_id);
CREATE INDEX idx_messages_status ON public.messages USING btree (status);
CREATE INDEX idx_messages_tenant ON public.messages USING btree (tenant_id);
CREATE INDEX idx_messages_tenant_contact ON public.messages USING btree (tenant_id, contact_id);
CREATE INDEX idx_messages_tenant_session ON public.messages USING btree (tenant_id, session_id);
CREATE INDEX idx_messages_tenant_timestamp ON public.messages USING btree (tenant_id, "timestamp");
CREATE INDEX idx_messages_timestamp ON public.messages USING btree ("timestamp");
CREATE INDEX idx_messages_updated ON public.messages USING btree (updated_at);
CREATE INDEX idx_messages_user ON public.messages USING btree (user_id);
CREATE INDEX idx_notes_author ON public.notes USING btree (author_id);
CREATE INDEX idx_notes_author_type ON public.notes USING btree (author_type);
CREATE INDEX idx_notes_contact ON public.notes USING btree (contact_id);
CREATE INDEX idx_notes_created ON public.notes USING btree (created_at);
CREATE INDEX idx_notes_deleted ON public.notes USING btree (deleted_at);
CREATE INDEX idx_notes_mentions ON public.notes USING gin (mentions);
CREATE INDEX idx_notes_pinned ON public.notes USING btree (pinned);
CREATE INDEX idx_notes_priority ON public.notes USING btree (priority);
CREATE INDEX idx_notes_session ON public.notes USING btree (session_id);
CREATE INDEX idx_notes_tags ON public.notes USING gin (tags);
CREATE INDEX idx_notes_tenant ON public.notes USING btree (tenant_id);
CREATE INDEX idx_notes_tenant_contact ON public.notes USING btree (tenant_id, contact_id);
CREATE INDEX idx_notes_tenant_priority ON public.notes USING btree (tenant_id, priority);
CREATE INDEX idx_notes_tenant_type ON public.notes USING btree (tenant_id, note_type);
CREATE INDEX idx_notes_type ON public.notes USING btree (note_type);
CREATE INDEX idx_notes_updated ON public.notes USING btree (updated_at);
CREATE INDEX idx_notes_visible ON public.notes USING btree (visible_to_client);
CREATE INDEX idx_outbox_aggregate ON public.outbox_events USING btree (aggregate_id, aggregate_type);
CREATE INDEX idx_outbox_correlation_id ON public.outbox_events USING gin (metadata);
CREATE INDEX idx_outbox_event_type ON public.outbox_events USING btree (event_type);
CREATE INDEX idx_outbox_events_deleted_at ON public.outbox_events USING btree (deleted_at);
CREATE UNIQUE INDEX idx_outbox_events_event_id ON public.outbox_events USING btree (event_id);
CREATE INDEX idx_outbox_retry ON public.outbox_events USING btree (retry_count, last_retry_at);
CREATE INDEX idx_outbox_status_created ON public.outbox_events USING btree (processed_at, status);
CREATE INDEX idx_outbox_tenant ON public.outbox_events USING btree (tenant_id);
CREATE INDEX idx_pipeline_statuses_active ON public.pipeline_statuses USING btree (active);
CREATE INDEX idx_pipeline_statuses_deleted_at ON public.pipeline_statuses USING btree (deleted_at);
CREATE INDEX idx_pipeline_statuses_pipeline_id ON public.pipeline_statuses USING btree (pipeline_id);
CREATE INDEX idx_pipeline_statuses_position ON public.pipeline_statuses USING btree ("position");
CREATE INDEX idx_pipeline_statuses_status_type ON public.pipeline_statuses USING btree (status_type);
CREATE INDEX idx_pipelines_active ON public.pipelines USING btree (active);
CREATE INDEX idx_pipelines_ai_provider ON public.pipelines USING btree (a_iprovider);
CREATE INDEX idx_pipelines_ai_summary ON public.pipelines USING btree (enable_ai_summary);
CREATE INDEX idx_pipelines_color ON public.pipelines USING btree (color);
CREATE INDEX idx_pipelines_created ON public.pipelines USING btree (created_at);
CREATE INDEX idx_pipelines_deleted ON public.pipelines USING btree (deleted_at);
CREATE INDEX idx_pipelines_name ON public.pipelines USING btree (name);
CREATE INDEX idx_pipelines_position ON public.pipelines USING btree ("position");
CREATE INDEX idx_pipelines_project ON public.pipelines USING btree (project_id);
CREATE INDEX idx_pipelines_tenant ON public.pipelines USING btree (tenant_id);
CREATE INDEX idx_pipelines_tenant_active ON public.pipelines USING btree (tenant_id, active);
CREATE INDEX idx_pipelines_tenant_name ON public.pipelines USING btree (tenant_id, name);
CREATE INDEX idx_pipelines_timeout ON public.pipelines USING btree (session_timeout_minutes);
CREATE INDEX idx_pipelines_updated ON public.pipelines USING btree (updated_at);
CREATE INDEX idx_processed_events_cleanup ON public.processed_events USING btree (processed_at);
CREATE INDEX idx_processed_events_lookup ON public.processed_events USING btree (consumer_name);
CREATE INDEX idx_projects_active ON public.projects USING btree (active);
CREATE INDEX idx_projects_billing ON public.projects USING btree (billing_account_id);
CREATE INDEX idx_projects_config ON public.projects USING gin (configuration);
CREATE INDEX idx_projects_created ON public.projects USING btree (created_at);
CREATE INDEX idx_projects_deleted ON public.projects USING btree (deleted_at);
CREATE INDEX idx_projects_name ON public.projects USING btree (name);
CREATE INDEX idx_projects_tenant ON public.projects USING btree (tenant_id);
CREATE INDEX idx_projects_tenant_active ON public.projects USING btree (tenant_id, active);
CREATE UNIQUE INDEX idx_projects_tenant_unique ON public.projects USING btree (tenant_id);
CREATE INDEX idx_projects_timeout ON public.projects USING btree (session_timeout_minutes);
CREATE INDEX idx_projects_updated ON public.projects USING btree (updated_at);
CREATE INDEX idx_projects_user ON public.projects USING btree (user_id);
CREATE INDEX idx_sessions_agent_ids ON public.sessions USING gin (agent_ids);
CREATE INDEX idx_sessions_channel_type ON public.sessions USING btree (channel_type_id);
CREATE INDEX idx_sessions_contact ON public.sessions USING btree (contact_id);
CREATE INDEX idx_sessions_converted ON public.sessions USING btree (converted);
CREATE INDEX idx_sessions_created ON public.sessions USING btree (created_at);
CREATE INDEX idx_sessions_deleted ON public.sessions USING btree (deleted_at);
CREATE INDEX idx_sessions_ended ON public.sessions USING btree (ended_at);
CREATE INDEX idx_sessions_escalated ON public.sessions USING btree (escalated);
CREATE INDEX idx_sessions_key_entities ON public.sessions USING gin (key_entities);
CREATE INDEX idx_sessions_last_activity ON public.sessions USING btree (last_activity_at);
CREATE INDEX idx_sessions_outcome_tags ON public.sessions USING gin (outcome_tags);
CREATE INDEX idx_sessions_pipeline ON public.sessions USING btree (pipeline_id);
CREATE INDEX idx_sessions_resolved ON public.sessions USING btree (resolved);
CREATE INDEX idx_sessions_sentiment ON public.sessions USING btree (sentiment);
CREATE INDEX idx_sessions_started ON public.sessions USING btree (started_at);
CREATE INDEX idx_sessions_status ON public.sessions USING btree (status);
CREATE INDEX idx_sessions_tenant ON public.sessions USING btree (tenant_id);
CREATE INDEX idx_sessions_tenant_contact ON public.sessions USING btree (tenant_id, contact_id);
CREATE INDEX idx_sessions_tenant_started ON public.sessions USING btree (tenant_id, started_at);
CREATE INDEX idx_sessions_tenant_status ON public.sessions USING btree (tenant_id, status);
CREATE INDEX idx_sessions_topics ON public.sessions USING gin (topics);
CREATE INDEX idx_sessions_updated ON public.sessions USING btree (updated_at);
CREATE INDEX idx_tracking_enrichments_deleted_at ON public.tracking_enrichments USING btree (deleted_at);
CREATE INDEX idx_trackings_ad_id ON public.trackings USING btree (ad_id);
CREATE INDEX idx_trackings_campaign ON public.trackings USING btree (campaign);
CREATE UNIQUE INDEX idx_trackings_click_id ON public.trackings USING btree (click_id);
CREATE INDEX idx_trackings_contact_id ON public.trackings USING btree (contact_id);
CREATE INDEX idx_trackings_created_at ON public.trackings USING btree (created_at);
CREATE INDEX idx_trackings_deleted_at ON public.trackings USING btree (deleted_at);
CREATE INDEX idx_trackings_platform ON public.trackings USING btree (platform);
CREATE INDEX idx_trackings_project_id ON public.trackings USING btree (project_id);
CREATE INDEX idx_trackings_session_id ON public.trackings USING btree (session_id);
CREATE INDEX idx_trackings_source ON public.trackings USING btree (source);
CREATE INDEX idx_trackings_tenant_id ON public.trackings USING btree (tenant_id);
CREATE INDEX idx_user_api_keys_active ON public.user_api_keys USING btree (active);
CREATE INDEX idx_user_api_keys_deleted_at ON public.user_api_keys USING btree (deleted_at);
CREATE UNIQUE INDEX idx_user_api_keys_key_hash ON public.user_api_keys USING btree (key_hash);
CREATE INDEX idx_user_api_keys_user_id ON public.user_api_keys USING btree (user_id);
CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);
CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);
CREATE INDEX idx_webhook_subscriptions_active ON public.webhook_subscriptions USING btree (active);
CREATE INDEX idx_webhook_subscriptions_deleted_at ON public.webhook_subscriptions USING btree (deleted_at);
CREATE INDEX idx_webhook_subscriptions_project_id ON public.webhook_subscriptions USING btree (project_id);
CREATE INDEX idx_webhook_subscriptions_subscribe_contact_events ON public.webhook_subscriptions USING btree (subscribe_contact_events);
CREATE INDEX idx_webhook_subscriptions_tenant_id ON public.webhook_subscriptions USING btree (tenant_id);
CREATE INDEX idx_webhook_subscriptions_user_id ON public.webhook_subscriptions USING btree (user_id);
CREATE UNIQUE INDEX unique_aggregate_sequence ON public.contact_event_store USING btree (sequence_number);
CREATE UNIQUE INDEX unique_aggregate_snapshot ON public.contact_snapshots USING btree (last_sequence_number);
CREATE UNIQUE INDEX uq_chats_external_id ON public.chats USING btree (external_id);
CREATE UNIQUE INDEX uq_processed_event_consumer ON public.processed_events USING btree (event_id, consumer_name);
ALTER TABLE ONLY public.agent_ai_interactions
    ADD CONSTRAINT fk_agent_ai_interactions_channel FOREIGN KEY (channel_id) REFERENCES public.channels(id);
ALTER TABLE ONLY public.agent_ai_interactions
    ADD CONSTRAINT fk_agent_ai_interactions_contact FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.agent_ai_interactions
    ADD CONSTRAINT fk_agent_ai_interactions_message_group FOREIGN KEY (group_id) REFERENCES public.message_groups(id);
ALTER TABLE ONLY public.agent_ai_interactions
    ADD CONSTRAINT fk_agent_ai_interactions_session FOREIGN KEY (session_id) REFERENCES public.sessions(id);
ALTER TABLE ONLY public.agent_sessions
    ADD CONSTRAINT fk_agent_sessions_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id);
ALTER TABLE ONLY public.agent_sessions
    ADD CONSTRAINT fk_agent_sessions_session FOREIGN KEY (session_id) REFERENCES public.sessions(id);
ALTER TABLE ONLY public.agents
    ADD CONSTRAINT fk_agents_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.agents
    ADD CONSTRAINT fk_agents_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.automations
    ADD CONSTRAINT fk_automations_pipeline FOREIGN KEY (pipeline_id) REFERENCES public.pipelines(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_billing_accounts_projects FOREIGN KEY (billing_account_id) REFERENCES public.billing_accounts(id);
ALTER TABLE ONLY public.billing_accounts
    ADD CONSTRAINT fk_billing_accounts_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.channels
    ADD CONSTRAINT fk_channels_pipeline FOREIGN KEY (pipeline_id) REFERENCES public.pipelines(id) ON DELETE SET NULL;
ALTER TABLE ONLY public.channels
    ADD CONSTRAINT fk_channels_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.channels
    ADD CONSTRAINT fk_channels_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_chats_messages FOREIGN KEY (chat_id) REFERENCES public.chats(id);
ALTER TABLE ONLY public.chats
    ADD CONSTRAINT fk_chats_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.contact_events
    ADD CONSTRAINT fk_contact_events_contact FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.contact_events
    ADD CONSTRAINT fk_contact_events_session FOREIGN KEY (session_id) REFERENCES public.sessions(id);
ALTER TABLE ONLY public.contact_pipeline_statuses
    ADD CONSTRAINT fk_contact_pipeline_statuses_contact FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.contact_pipeline_statuses
    ADD CONSTRAINT fk_contact_pipeline_statuses_pipeline FOREIGN KEY (pipeline_id) REFERENCES public.pipelines(id);
ALTER TABLE ONLY public.contact_pipeline_statuses
    ADD CONSTRAINT fk_contact_pipeline_statuses_status FOREIGN KEY (status_id) REFERENCES public.pipeline_statuses(id);
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_contacts_messages FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT fk_contacts_sessions FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.message_enrichments
    ADD CONSTRAINT fk_message_enrichments_message FOREIGN KEY (message_id) REFERENCES public.messages(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.message_enrichments
    ADD CONSTRAINT fk_message_enrichments_message_group FOREIGN KEY (message_group_id) REFERENCES public.message_groups(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_messages_channel FOREIGN KEY (channel_id) REFERENCES public.channels(id) ON DELETE RESTRICT;
ALTER TABLE ONLY public.notes
    ADD CONSTRAINT fk_notes_contact FOREIGN KEY (contact_id) REFERENCES public.contacts(id);
ALTER TABLE ONLY public.notes
    ADD CONSTRAINT fk_notes_session FOREIGN KEY (session_id) REFERENCES public.sessions(id);
ALTER TABLE ONLY public.pipeline_statuses
    ADD CONSTRAINT fk_pipeline_statuses_pipeline FOREIGN KEY (pipeline_id) REFERENCES public.pipelines(id);
ALTER TABLE ONLY public.contacts
    ADD CONSTRAINT fk_projects_contacts FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_projects_messages FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.pipelines
    ADD CONSTRAINT fk_projects_pipelines FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.messages
    ADD CONSTRAINT fk_sessions_messages FOREIGN KEY (session_id) REFERENCES public.sessions(id);
ALTER TABLE ONLY public.user_api_keys
    ADD CONSTRAINT fk_users_api_keys FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_users_projects FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.webhook_subscriptions
    ADD CONSTRAINT fk_webhook_subscriptions_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.webhook_subscriptions
    ADD CONSTRAINT fk_webhook_subscriptions_user FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
