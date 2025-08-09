#!/bin/bash

echo "🚀 Go API Server 초기 설정 스크립트"
echo "====================================="

# 1. .env 파일 생성
if [ ! -f .env ]; then
    echo "📝 .env 파일 생성 중..."
    cp .env.example .env
    echo "✅ .env 파일이 생성되었습니다. DB 정보를 수정해주세요."
else
    echo "ℹ️  .env 파일이 이미 존재합니다."
fi

# 2. Go 모듈 다운로드
echo ""
echo "📦 Go 모듈 다운로드 중..."
go mod download
go mod tidy
echo "✅ Go 모듈 다운로드 완료"

# 3. 데이터베이스 타입 확인
echo ""
echo "🗄️  어떤 데이터베이스를 사용하시겠습니까?"
echo "1) PostgreSQL (권장)"
echo "2) Oracle"
read -p "선택 (1 또는 2): " db_choice

if [ "$db_choice" = "1" ]; then
    echo ""
    echo "PostgreSQL 설정:"
    echo "----------------"
    
    # PostgreSQL이 설치되어 있는지 확인
    if command -v psql &> /dev/null; then
        echo "✅ PostgreSQL이 설치되어 있습니다."
        
        # 데이터베이스 생성 제안
        read -p "praytogether 데이터베이스를 생성하시겠습니까? (y/n): " create_db
        if [ "$create_db" = "y" ]; then
            createdb praytogether
            echo "✅ 데이터베이스가 생성되었습니다."
        fi
    else
        echo "⚠️  PostgreSQL이 설치되어 있지 않습니다."
        echo "다음 명령으로 설치하세요:"
        echo "  macOS: brew install postgresql"
        echo "  Ubuntu: sudo apt-get install postgresql"
    fi
    
    # .env 파일 업데이트
    sed -i.bak 's/DB_TYPE=.*/DB_TYPE=postgres/' .env
    echo "✅ .env 파일이 PostgreSQL용으로 설정되었습니다."
    
elif [ "$db_choice" = "2" ]; then
    echo ""
    echo "Oracle 설정:"
    echo "-----------"
    echo "⚠️  Oracle Instant Client가 필요합니다."
    echo "설치되어 있지 않다면 Oracle 웹사이트에서 다운로드하세요."
    
    # .env 파일 업데이트
    sed -i.bak 's/DB_TYPE=.*/DB_TYPE=oracle/' .env
    echo "✅ .env 파일이 Oracle용으로 설정되었습니다."
fi

# 4. Air 설치 제안
echo ""
echo "🔄 Hot Reload를 위한 Air 설치를 권장합니다."
read -p "Air를 설치하시겠습니까? (y/n): " install_air
if [ "$install_air" = "y" ]; then
    go install github.com/cosmtrek/air@latest
    echo "✅ Air가 설치되었습니다."
    echo "실행: air"
fi

echo ""
echo "====================================="
echo "✅ 설정이 완료되었습니다!"
echo ""
echo "다음 명령으로 서버를 실행하세요:"
echo "  make run"
echo "또는"
echo "  go run cmd/api/main.go"
echo ""
echo "Hot Reload (Air 설치 시):"
echo "  air"
echo ""
echo "API 테스트:"
echo "  curl http://localhost:8080/health"
echo "====================================="