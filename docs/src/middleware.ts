import type { MiddlewareHandler } from "astro";

export const onRequest: MiddlewareHandler = async (_, next) => {
	const response = await next();

	// Security Headers for SEO and security
	response.headers.set("X-Frame-Options", "DENY");
	response.headers.set("X-XSS-Protection", "1; mode=block");
	response.headers.set("X-Content-Type-Options", "nosniff");

	// Referrer Policy
	response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
	response.headers.set(
		"Strict-Transport-Security",
		"max-age=31536000; includeSubDomains; preload",
	);

	// Content Security Policy
	const csp = [
		"base-uri 'self'",
		"object-src 'none'",
		"default-src 'self'",
		"connect-src 'self'",
		"frame-ancestors 'none'",
		"img-src 'self' data: https:",
		"font-src 'self' https://fonts.gstatic.com",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
		"script-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
	].join("; ");

	response.headers.set("Content-Security-Policy", csp);

	return response;
};
