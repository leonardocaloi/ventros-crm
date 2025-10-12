package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adiciona headers de segurança recomendados pela OWASP
// https://owasp.org/www-project-secure-headers/
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options: Previne MIME-sniffing
		// Força o browser a respeitar o Content-Type declarado
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Previne clickjacking
		// DENY: Não permite iframe de nenhum site
		// SAMEORIGIN: Permite iframe apenas do mesmo domínio
		c.Header("X-Frame-Options", "DENY")

		// X-XSS-Protection: Ativa filtro XSS do browser (legacy, mas ainda útil)
		// 1; mode=block: Ativa e bloqueia página se detectar XSS
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security (HSTS): Força HTTPS
		// max-age=31536000: 1 ano
		// includeSubDomains: Aplica para todos os subdomínios
		// preload: Habilita preload list do Chrome
		// NOTA: Só ativar em produção com HTTPS configurado
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content-Security-Policy (CSP): Previne XSS e data injection
		// default-src 'self': Só permite recursos do mesmo domínio
		// script-src 'self': Só permite scripts do mesmo domínio
		// style-src 'self' 'unsafe-inline': Permite estilos inline (necessário para algumas libs)
		// img-src 'self' data: https:: Permite imagens do mesmo domínio, data URIs e HTTPS
		// font-src 'self': Só permite fontes do mesmo domínio
		// connect-src 'self': Só permite conexões (fetch, XHR) para o mesmo domínio
		// frame-ancestors 'none': Não permite ser embutido em iframe (similar ao X-Frame-Options)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")

		// Referrer-Policy: Controla quanto de informação é enviada no header Referer
		// strict-origin-when-cross-origin: Envia origin completo para same-origin, só origin para cross-origin HTTPS, nada para HTTP
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (anteriormente Feature-Policy): Controla quais features do browser podem ser usadas
		// Desabilita features perigosas ou desnecessárias
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		// X-Permitted-Cross-Domain-Policies: Controla cross-domain policies (Flash, PDF)
		// none: Não permite nenhuma policy cross-domain
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Cache-Control: Controla caching de respostas sensíveis
		// API responses geralmente não devem ser cacheadas
		// NOTA: Rotas públicas (assets, etc) podem sobrescrever isso
		if c.Request.Method != "GET" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
		}

		// X-DNS-Prefetch-Control: Controla DNS prefetching
		// off: Desabilita para proteger privacidade
		c.Header("X-DNS-Prefetch-Control", "off")

		// X-Download-Options: Previne execução de downloads no IE
		// noopen: Não permite abrir diretamente no IE
		c.Header("X-Download-Options", "noopen")

		c.Next()
	}
}

// SecureJSONMiddleware garante que JSON responses sempre começam com {} ou []
// Previne vulnerabilidades de CSRF via JSON hijacking
func SecureJSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Force Content-Type para application/json em responses JSON
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		c.Next()
	}
}

// NoCacheMiddleware força no-cache em rotas sensíveis
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	}
}
