# Control D Sync

Utilitário em Go que mantém suas pastas do Control D sincronizadas com listas de bloqueio remotas.

## O que faz

Este utilitário faz três coisas principais:
1. **Lê os nomes das pastas** dos arquivos JSON remotos
2. **Deleta pastas existentes** com esses nomes (para começar do zero)
3. **Recria as pastas** e adiciona todas as regras em lotes

Nada complicado, apenas funciona.

## Características

- **Performance superior**: Binário compilado nativo
- **Deploy simples**: Binário único sem dependências externas
- **Eficiente**: Baixo uso de memória (~10-15MB)
- **Rápido**: Startup em ~100ms

## Configuração

### 1. Variáveis de ambiente

Crie um arquivo `.env` baseado no `.env.example`:

```bash
TOKEN=seu_token_do_control_d_aqui
PROFILE=seu_profile_id_aqui
```

- `TOKEN`: Seu token de API do Control D
- `PROFILE`: ID do perfil (ou múltiplos IDs separados por vírgula)

### 2. Compilação

```bash
# Instalar dependências
go mod tidy

# Compilar o binário
go build -o ctrld-sync main.go
```

### 3. Execução

```bash
# Executar diretamente
./ctrld-sync

# Ou executar com go run
go run main.go
```

## Funcionalidades

### Listas de bloqueio suportadas

O script sincroniza automaticamente com as seguintes listas do [hagezi/dns-blocklists](https://github.com/hagezi/dns-blocklists):

- Apple Private Relay Allow
- Native Tracker (Amazon, Apple, Huawei, LG WebOS, Microsoft, OPPO/Realme, Roku, Samsung, TikTok, Vivo, Xiaomi)
- Ultimate Known Issues Allow
- Referral Allow
- Spam (IDNs, TLDs, TLDs Allow)
- Badware Hoster

### Características técnicas

- **Processamento concorrente**: Múltiplos perfis sincronizados simultaneamente (máx. 3)
- **Retry logic**: Tentativas automáticas com backoff exponencial
- **Processamento em lotes**: Regras enviadas em grupos de 500
- **Cache inteligente**: URLs já buscadas são mantidas em cache
- **Detecção de duplicatas**: Evita regras duplicadas entre pastas
- **Logging detalhado**: Acompanhe o progresso em tempo real
- **Múltiplos perfis**: Suporte a vários perfis Control D com sincronização paralela

## Estrutura do código

```
main.go          # Código principal
go.mod           # Dependências Go
go.sum           # Lock file das dependências
.env.example     # Exemplo de configuração
README-go.md     # Esta documentação
```

## Processamento Concorrente

### Como funciona

Quando você tem múltiplos perfis configurados (separados por vírgula), o script processa até **3 perfis simultaneamente** usando goroutines:

```bash
# Exemplo com múltiplos perfis
PROFILE=profile1,profile2,profile3,profile4,profile5
```

### Benefícios da concorrência

- **3-5x mais rápido** para múltiplos perfis
- **Uso eficiente** de recursos de rede
- **Controle de limite** para não sobrecarregar a API
- **Processamento independente** - falha em um perfil não afeta outros

### Logs de exemplo (múltiplos perfis)

```
Starting concurrent sync for 3 profiles (max 3 concurrent)
Starting sync for profile 12345
Starting sync for profile 67890
Starting sync for profile 54321
Deleted folder 'Apple Private Relay Allow' (ID 67890)
Created folder 'Apple Private Relay Allow' (ID 67891)
Folder 'Apple Private Relay Allow' – batch 1: added 500 rules
Folder 'Apple Private Relay Allow' – finished (1247 new rules added)
Sync complete: 18/18 folders processed successfully
All profiles processed: 3/3 successful
```

## Logs de exemplo (perfil único)

```
Starting concurrent sync for 1 profiles (max 3 concurrent)
Starting sync for profile 12345
Deleted folder 'Apple Private Relay Allow' (ID 67890)
Created folder 'Apple Private Relay Allow' (ID 67891)
Folder 'Apple Private Relay Allow' – batch 1: added 500 rules
Folder 'Apple Private Relay Allow' – finished (1247 new rules added)
Sync complete: 18/18 folders processed successfully
All profiles processed: 1/1 successful
```

## Troubleshooting

### Erro de compilação
```bash
# Limpar cache e reinstalar dependências
go clean -modcache
go mod tidy
```

### Erro de permissão
```bash
# Dar permissão de execução ao binário
chmod +x ctrld-sync
```

### Problemas de API
- Verifique se o TOKEN está correto
- Confirme se o PROFILE ID existe
- Verifique sua conexão com a internet

## Desenvolvimento

### Estrutura do código Go

- **Structs**: Definições de tipos para JSON da API
- **HTTP Clients**: Clientes separados para API e GitHub
- **Retry Logic**: Implementação robusta com backoff exponencial
- **Error Handling**: Tratamento de erros detalhado
- **Logging**: Sistema de logs estruturado
- **Concorrência**: Goroutines com semáforos para controle de limite

### Executar em modo debug

```bash
# Compilar com informações de debug
go build -gcflags="all=-N -l" -o ctrld-sync-debug main.go

# Executar com logs verbosos
./ctrld-sync-debug
```

### Configurações avançadas

Para ajustar o limite de concorrência, modifique a constante no código:

```go
const MaxConcurrentProfiles = 3 // Ajuste conforme necessário
```

## Performance

### Métricas típicas

- **Tempo de startup**: ~100ms
- **Uso de memória**: ~10-15MB
- **Tamanho do executável**: ~8-12MB
- **Dependências externas**: Nenhuma

### Otimizações implementadas

- Cache em memória para URLs já buscadas
- Processamento concorrente de múltiplos perfis
- Retry logic com backoff exponencial
- Detecção de duplicatas para evitar regras redundantes

## Licença

MIT License - veja o arquivo LICENSE para detalhes.
