# Projeto: Microservices with Go

Este é o código inicial do curso **"Microservices with Go"**, onde construímos o backend de um aplicativo estilo Uber, usando **Go**, **Docker** e **Kubernetes**.

No final do projeto, teremos um sistema de ride-sharing completo, com vários microsserviços rodando em containers e orquestrados no Kubernetes, pronto para escalar horizontalmente e receber tráfego real.

---

## Visão geral

A ideia do projeto é desenvolver passo a passo os serviços que compõem um app de transporte (tipo Uber).
Durante o curso, vamos aprender a criar e rodar tudo localmente com Docker e Kubernetes.

---

## Requisitos

Antes de começar, é preciso ter instalado:

- **Docker**
- **Go**
- **Tilt** (para rodar o ambiente local)
- **Kubernetes** (Via Minikube)

---

## Instalação (MacOS)

1. Instale o **Homebrew**
   https://brew.sh/

2. Instale o **Docker Desktop**
   https://www.docker.com/products/docker-desktop/

3. Instale o **Minikube**
   https://minikube.sigs.k8s.io/docs/

4. Instale o **Tilt**
   https://tilt.dev/

5. Instale o **Go** com o Homebrew:
   ```bash
   brew install go
   ```

6. Instale o **kubectl**:
   https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/

---

## Instalação (Windows via WSL)

1. Instale o **WSL**:  
   https://learn.microsoft.com/en-us/windows/wsl/install

2. Instale o **Docker Desktop**  
   https://www.docker.com/products/docker-desktop/

3. Instale o **Minikube**  
   https://minikube.sigs.k8s.io/docs/

4. Instale o **Tilt**  
   https://tilt.dev/

5. Instale o **Go** dentro do WSL:
   ```bash
   # Baixar o binário
   wget https://dl.google.com/go/go1.23.0.linux-amd64.tar.gz

   # Extrair o arquivo
   sudo tar -xvf go1.23.0.linux-amd64.tar.gz

   # Mover para /usr/local
   sudo mv go /usr/local

   # Adicionar o Go ao PATH
   echo 'export GOROOT=/usr/local/go' >> ~/.bashrc
   echo 'export GOPATH=$HOME/go' >> ~/.bashrc
   echo 'export PATH=$GOPATH/bin:$GOROOT/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc

   # Verificar instalação
   go version
   ```

6. Instale o **kubectl**:  
   https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/

---

## Rodando o projeto

Depois de tudo instalado e configurado, você pode iniciar o ambiente de desenvolvimento com:

```bash
tilt up
```

Isso vai construir os containers e iniciar os serviços dentro do seu cluster local do Kubernetes.

---

## Monitorando os serviços

Para ver se os pods estão rodando corretamente:

```bash
kubectl get pods
```