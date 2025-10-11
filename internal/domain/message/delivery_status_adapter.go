package message

// DeliveryStatusAdapter é a interface que todo provedor de canal deve implementar
// para traduzir seus status específicos para o Status agnóstico do domínio.
//
// Cada infrastructure/channel implementa seu próprio adapter:
// - WAHA: waha.WhatsAppAck implementa esta interface
// - Telegram: telegram.UpdateStatus implementaria esta interface
// - SMS: sms.DeliveryReceipt implementaria esta interface
//
// Exemplo de uso:
//
//	// Infrastructure layer
//	wahaAck := waha.NewWhatsAppAck(2)  // ACK 2 = DEVICE
//
//	// Conversão via interface
//	var adapter message.DeliveryStatusAdapter = wahaAck
//	domainStatus, err := adapter.ToMessageStatus()
//	// domainStatus == message.StatusDelivered
type DeliveryStatusAdapter interface {
	// ToMessageStatus converte o status específico do provedor
	// para o Status agnóstico do domínio.
	ToMessageStatus() (Status, error)

	// IsValid verifica se o status do provedor é válido.
	IsValid() bool

	// String retorna a representação em string do status do provedor.
	String() string
}
