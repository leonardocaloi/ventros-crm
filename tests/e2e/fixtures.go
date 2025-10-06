package e2e

// TestFixtures contém os dados de entrada para os testes
type TestFixtures struct {
	Users    []UserFixture
	Channels []ChannelFixture
	Contacts []ContactFixture
}

type UserFixture struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type ChannelFixture struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	WAHAConfig map[string]string `json:"waha_config,omitempty"`
}

type ContactFixture struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// GetDefaultFixtures retorna os dados de teste padrão
func GetDefaultFixtures() *TestFixtures {
	return &TestFixtures{
		Users: []UserFixture{
			{
				Name:     "Admin Teste",
				Email:    "admin.teste@ventros.local",
				Password: "senha_teste_123",
				Role:     "admin",
			},
			{
				Name:     "Vendedor Teste",
				Email:    "vendedor.teste@ventros.local",
				Password: "senha_teste_123",
				Role:     "user",
			},
		},
		Channels: []ChannelFixture{
			{
				Name: "WhatsApp Teste",
				Type: "waha",
				WAHAConfig: map[string]string{
					"base_url":    "https://waha.ventros.cloud",
					"token":       "test_token_123",
					"session_id":  "test_session",
					"webhook_url": "http://localhost:8080/api/v1/webhooks/waha",
				},
			},
		},
		Contacts: []ContactFixture{
			{
				Name:  "João Silva Teste",
				Phone: "+5511999887766",
				Email: "joao.teste@example.com",
			},
			{
				Name:  "Maria Santos Teste",
				Phone: "+5511988776655",
				Email: "maria.teste@example.com",
			},
		},
	}
}
