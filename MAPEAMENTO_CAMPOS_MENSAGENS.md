# Mapeamento de Campos: WAHA → Tabela Messages

## 📊 Visão Geral

Este documento mapeia **exatamente** quais campos do payload WAHA são salvos em quais colunas da tabela `messages` para cada tipo de mensagem.

---

## 🗄️ Tabela `messages` - Schema

```sql
CREATE TABLE messages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp           TIMESTAMP NOT NULL,
    user_id             UUID NOT NULL,           -- Dono do workspace
    project_id          UUID NOT NULL,
    channel_type_id     INT,                     -- 1 = WhatsApp
    from_me             BOOLEAN DEFAULT false,
    channel_id          UUID NOT NULL,           -- Canal específico
    contact_id          UUID NOT NULL,
    session_id          UUID,                    -- Sessão ativa
    content_type        VARCHAR NOT NULL,        -- text/image/audio/voice/video/document/location/contact/sticker
    text                TEXT,                    -- Conteúdo textual
    media_url           VARCHAR,                 -- URL da mídia
    media_mimetype      VARCHAR,                 -- Tipo MIME
    channel_message_id  VARCHAR,                 -- ID externo
    reply_to_id         UUID,                    -- Resposta a outra mensagem
    status              VARCHAR DEFAULT 'sent',  -- sent/delivered/read/failed
    language            VARCHAR,
    agent_id            UUID,
    metadata            JSONB,                   -- Dados extras
    delivered_at        TIMESTAMP,
    read_at             TIMESTAMP,
    created_at          TIMESTAMP DEFAULT NOW(),
    updated_at          TIMESTAMP DEFAULT NOW(),
    deleted_at          TIMESTAMP
);
```

---

## 📝 1. TEXT

### Exemplo JSON
```json
{
  "payload": {
    "id": "false_554497044474@c.us_3F0B3ABFCA9801F3A48F",
    "timestamp": 1759875205,
    "from": "554497044474@c.us",
    "fromMe": false,
    "body": "Teste",
    "_data": {
      "Info": {
        "Type": "text",
        "PushName": "Leonardo"
      },
      "Message": {
        "extendedTextMessage": {
          "text": "Teste"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"text"` | `"text"` |
| `text` | `payload.body` ou `payload._data.Message.extendedTextMessage.text` | `"Teste"` |
| `channel_message_id` | `payload.id` | `"false_554497044474@c.us_3F0B..."` |
| `timestamp` | `payload.timestamp` (unix) | `2025-10-07 19:13:25` |
| `metadata` | Vários | `{"waha_event_id": "evt_...", "source": "app"}` |

---

## 🖼️ 2. IMAGE

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../image.jpeg",
      "mimetype": "image/jpeg"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "image"
      },
      "Message": {
        "imageMessage": {
          "caption": "Legenda da foto"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"image"` | `"image"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../image.jpeg"` |
| `media_mimetype` | `payload.media.mimetype` | `"image/jpeg"` |
| `text` | `payload._data.Message.imageMessage.caption` | `"Legenda da foto"` (se houver) |

---

## 🔊 3. AUDIO

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../audio.oga",
      "mimetype": "audio/ogg; codecs=opus"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "audio"
      },
      "Message": {
        "audioMessage": {
          "PTT": false
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"audio"` | `"audio"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../audio.oga"` |
| `media_mimetype` | `payload.media.mimetype` | `"audio/ogg; codecs=opus"` |

---

## 🎤 4. VOICE (PTT)

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../voice.oga",
      "mimetype": "audio/ogg; codecs=opus"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "ptt"  ← DIFERENÇA
      },
      "Message": {
        "audioMessage": {
          "PTT": true  ← DIFERENÇA
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"voice"` | `"voice"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../voice.oga"` |
| `media_mimetype` | `payload.media.mimetype` | `"audio/ogg; codecs=opus"` |

**Diferença entre Audio e Voice:**
- Audio: `MediaType: "audio"` + `PTT: false`
- Voice: `MediaType: "ptt"` + `PTT: true`

---

## 🎥 5. VIDEO

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../video.mp4",
      "mimetype": "video/mp4"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "video"
      },
      "Message": {
        "videoMessage": {
          "caption": "Legenda do vídeo"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"video"` | `"video"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../video.mp4"` |
| `media_mimetype` | `payload.media.mimetype` | `"video/mp4"` |
| `text` | `payload._data.Message.videoMessage.caption` | `"Legenda do vídeo"` (se houver) |

---

## 📄 6. DOCUMENT

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../doc.pdf",
      "mimetype": "application/pdf",
      "filename": "DOC-20241112-WA0012..pdf"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "document"
      },
      "Message": {
        "documentMessage": {
          "fileName": "DOC-20241112-WA0012..pdf",
          "caption": "Documento importante"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"document"` | `"document"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../doc.pdf"` |
| `media_mimetype` | `payload.media.mimetype` | `"application/pdf"` |
| `text` | `payload._data.Message.documentMessage.caption` | `"Documento importante"` (se houver) |
| `metadata.filename` | `payload.media.filename` | `"DOC-20241112-WA0012..pdf"` |

**Tipos de documento suportados:**
- PDF: `application/pdf`
- HEIC: `image/heic`
- DOCX: `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- XLSX: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- Etc.

---

## 📍 7. LOCATION

### Exemplo JSON
```json
{
  "payload": {
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "location"
      },
      "Message": {
        "locationMessage": {
          "degreesLatitude": -23.408384323120117,
          "degreesLongitude": -51.939579010009766,
          "name": "Meu Local",
          "address": "Rua Exemplo, 123"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"location"` | `"location"` |
| `metadata.location.latitude` | `payload._data.Message.locationMessage.degreesLatitude` | `-23.408384323120117` |
| `metadata.location.longitude` | `payload._data.Message.locationMessage.degreesLongitude` | `-51.939579010009766` |
| `metadata.location.name` | `payload._data.Message.locationMessage.name` | `"Meu Local"` (opcional) |
| `metadata.location.address` | `payload._data.Message.locationMessage.address` | `"Rua Exemplo, 123"` (opcional) |

**Metadata JSON:**
```json
{
  "location": {
    "latitude": -23.408384323120117,
    "longitude": -51.939579010009766,
    "name": "Meu Local",
    "address": "Rua Exemplo, 123"
  }
}
```

---

## 👤 8. CONTACT (VCard)

### Exemplo JSON
```json
{
  "payload": {
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "vcard"
      },
      "Message": {
        "contactMessage": {
          "displayName": "Leonardo Caloi Santos",
          "vcard": "BEGIN:VCARD\nVERSION:3.0\nN:Santos;Leonardo Caloi;;;\nFN:Leonardo Caloi Santos\nTEL;type=CELL:+55 44 99704-4474\nEND:VCARD"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"contact"` | `"contact"` |
| `metadata.contact.display_name` | `payload._data.Message.contactMessage.displayName` | `"Leonardo Caloi Santos"` |
| `metadata.contact.vcard` | `payload._data.Message.contactMessage.vcard` | `"BEGIN:VCARD\n..."` |

**Metadata JSON:**
```json
{
  "contact": {
    "display_name": "Leonardo Caloi Santos",
    "vcard": "BEGIN:VCARD\nVERSION:3.0\nN:Santos;Leonardo Caloi;;;\nFN:Leonardo Caloi Santos\nTEL;type=CELL:+55 44 99704-4474\nEND:VCARD"
  }
}
```

---

## 🎨 9. STICKER

### Exemplo JSON
```json
{
  "payload": {
    "hasMedia": true,
    "media": {
      "url": "https://storage.googleapis.com/.../sticker.webp",
      "mimetype": "image/webp"
    },
    "_data": {
      "Info": {
        "Type": "media",
        "MediaType": "sticker"
      },
      "Message": {
        "stickerMessage": {
          "URL": "...",
          "mimetype": "image/webp"
        }
      }
    }
  }
}
```

### Mapeamento
| Campo DB | Fonte WAHA | Valor Exemplo |
|----------|------------|---------------|
| `content_type` | `"sticker"` | `"sticker"` |
| `media_url` | `payload.media.url` | `"https://storage.googleapis.com/.../sticker.webp"` |
| `media_mimetype` | `payload.media.mimetype` | `"image/webp"` |

---

## 🔗 Campos Comuns (Todos os Tipos)

### Mapeamento Universal
| Campo DB | Fonte WAHA | Descrição |
|----------|------------|-----------|
| `id` | Auto-gerado (UUID) | ID único da mensagem no CRM |
| `timestamp` | `payload.timestamp` (unix → datetime) | Data/hora da mensagem |
| `user_id` | `channel.user_id` | Dono do workspace |
| `project_id` | `channel.project_id` | Projeto |
| `channel_type_id` | AppConfig (1 = WhatsApp) | Tipo do canal |
| `from_me` | `payload.fromMe` | Enviada por mim? |
| `channel_id` | `channel.id` (lookup por `session`) | Canal específico |
| `contact_id` | Lookup/Create por `payload.from` | Contato |
| `session_id` | FindOrCreate por contact + canal | Sessão ativa |
| `channel_message_id` | `payload.id` | ID externo do WhatsApp |
| `status` | `"sent"` (default) | Status de entrega |

### Metadata Comum
```json
{
  "waha_event_id": "evt_01k70bs22nkd0r925e7yvp5xjj",
  "waha_session": "guilherme-batilani-suporte",
  "channel_id": "uuid-do-canal",
  "channel_name": "Suporte Guilherme",
  "source": "app",
  "is_from_ad": false
}
```

---

## 📊 Tracking de Conversão (Ads)

### Quando mensagem vem de anúncio

**Detectado por:**
```json
{
  "payload": {
    "_data": {
      "Message": {
        "extendedTextMessage": {
          "contextInfo": {
            "entryPointConversionSource": "ad",
            "entryPointConversionApp": "facebook",
            "externalAdReply": {
              "ctwaClid": "click-id-123"
            }
          }
        }
      }
    }
  }
}
```

**Metadata adicional:**
```json
{
  "is_from_ad": true,
  "tracking": {
    "conversion_source": "ad",
    "conversion_app": "facebook",
    "external_source": "instagram",
    "external_medium": "story",
    "ad_source_type": "ad",
    "ad_source_id": "123456",
    "ctwa_clid": "click-id-123"
  }
}
```

---

## 🔍 Queries Úteis

### Buscar mensagens por tipo
```sql
SELECT 
  id,
  content_type,
  text,
  media_url,
  metadata,
  created_at
FROM messages
WHERE content_type = 'voice'  -- ou 'location', 'contact', etc
ORDER BY created_at DESC
LIMIT 10;
```

### Buscar mensagens com localização
```sql
SELECT 
  id,
  metadata->'location'->>'latitude' AS lat,
  metadata->'location'->>'longitude' AS lng,
  metadata->'location'->>'name' AS local_name,
  created_at
FROM messages
WHERE content_type = 'location'
  AND metadata->'location' IS NOT NULL;
```

### Buscar mensagens de contato
```sql
SELECT 
  id,
  metadata->'contact'->>'display_name' AS nome,
  metadata->'contact'->>'vcard' AS vcard,
  created_at
FROM messages
WHERE content_type = 'contact';
```

### Buscar documentos com filename
```sql
SELECT 
  id,
  media_url,
  media_mimetype,
  metadata->>'filename' AS nome_arquivo,
  created_at
FROM messages
WHERE content_type = 'document'
  AND metadata->>'filename' IS NOT NULL;
```

### Buscar mensagens de ads
```sql
SELECT 
  id,
  text,
  metadata->>'is_from_ad' AS eh_de_ad,
  metadata->'tracking'->>'ctwa_clid' AS click_id,
  created_at
FROM messages
WHERE (metadata->>'is_from_ad')::boolean = true;
```

---

## 📈 Estatísticas

### Contagem por tipo
```sql
SELECT 
  content_type,
  COUNT(*) AS total,
  COUNT(CASE WHEN from_me THEN 1 END) AS enviadas,
  COUNT(CASE WHEN NOT from_me THEN 1 END) AS recebidas
FROM messages
GROUP BY content_type
ORDER BY total DESC;
```

### Exemplo de resultado:
```
content_type | total | enviadas | recebidas
-------------|-------|----------|----------
text         | 1523  | 892      | 631
image        | 342   | 156      | 186
voice        | 189   | 23       | 166
document     | 87    | 67       | 20
video        | 45    | 12       | 33
location     | 23    | 5        | 18
audio        | 12    | 8        | 4
contact      | 8     | 2        | 6
sticker      | 3     | 1        | 2
```

---

## ✅ Resumo

**Todos os tipos de mensagem são salvos na mesma tabela `messages`.**

**Campos variáveis por tipo:**
- `content_type` → Define o tipo
- `text` → Usado em text, captions
- `media_url` + `media_mimetype` → Usado em image/audio/video/document/sticker
- `metadata` → Dados específicos (location coords, contact vcard, filename)

**Campos fixos:**
- Identificação: `id`, `channel_message_id`
- Contexto: `user_id`, `project_id`, `channel_id`, `contact_id`, `session_id`
- Temporal: `timestamp`, `created_at`, `updated_at`
- Status: `status`, `from_me`
