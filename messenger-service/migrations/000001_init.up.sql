CREATE SCHEMA IF NOT EXISTS messengerservice;

CREATE TABLE IF NOT EXISTS messengerservice.conversations (
    id              UUID PRIMARY KEY,
    listing_id      UUID NOT NULL,
    seller_id       UUID NOT NULL,
    buyer_id        UUID NOT NULL CHECK (buyer_id <> seller_id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (listing_id, buyer_id)
);

CREATE TABLE IF NOT EXISTS messengerservice.messages (
    id              UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES messengerservice.conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL,
    body            TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_seller ON messengerservice.conversations(seller_id);
CREATE INDEX IF NOT EXISTS idx_conversations_buyer  ON messengerservice.conversations(buyer_id);
CREATE INDEX IF NOT EXISTS idx_messages_conv_created ON messengerservice.messages(conversation_id, created_at DESC);
