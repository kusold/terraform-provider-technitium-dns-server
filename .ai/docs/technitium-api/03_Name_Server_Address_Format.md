# Technitium DNS Server API - Name Server Address Format

## Name Server Address Format

The DNS server uses a specific text format to define the name server address to allow specifying multiple parameters like the domain name, IP address, port or URL. This format is used in the web console as well as in this API. It is used to specify forwarder address in DNS settings, conditional forwarder zone's FWD record, or the server address in DNS Client resolve query API calls.

- A name server address with just an IP address is specified just as its string literal with optional port number is as shown: `1.1.1.1` or `8.8.8.8:53`. When port is not specified, the default port number for the selected DNS transport protocol is used.
- A name server address with just a domain name is specified similarly as its string literal with optional port number is as shown: `dns.quad9.net:853` or `cloudflare-dns.com`. When port is not specified, the default port number for the selected DNS transport protocol is used.
- A combination of domain name and IP address together with optional port number is as shown: `cloudflare-dns.com (1.1.1.1)`, `dns.quad9.net (9.9.9.9:853)` or `dns.quad9.net:853 (9.9.9.9)`. Here, the domain name (with optional port number) is specified and the IP address (with optional port number) is specified in a round bracket. When port is not specified, the default port number for the selected DNS transport protocol is used. This allows the DNS server to use the specified IP address instead of trying to resolve it separately.
- A name server address that specifies a DNS-over-HTTPS URL is specified just as its string literal is as shown: `https://cloudflare-dns.com/dns-query`
- A combination of DNS-over-HTTPS URL and IP address together is as shown: `https://cloudflare-dns.com/dns-query (1.1.1.1)`. Here, the IP address of the domain name in the URL is specified in the round brackets. This allows the DNS server to use the specified IP address instead of trying to resolve it separately.
- IPv6 addresses must always be enclosed in square brackets when port is specified as shown: `cloudflare-dns.com ([2606:4700:4700::1111]:853)` or `[2606:4700:4700::1111]:853`
