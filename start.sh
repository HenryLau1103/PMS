#!/bin/bash

echo "🚀 啟動 PSM 台股智能投資組合管理系統"
echo "=========================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker 未運行，請先啟動 Docker"
    exit 1
fi

echo "✅ Docker 已就緒"
echo ""

# Check if docker-compose.yml exists
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ Error: docker-compose.yml 不存在"
    exit 1
fi

echo "📦 啟動服務..."
docker-compose up -d

echo ""
echo "⏳ 等待服務啟動 (約 30 秒)..."
sleep 30

echo ""
echo "🔍 檢查服務狀態..."
docker-compose ps

echo ""
echo "🏥 健康檢查..."
HEALTH_CHECK=$(curl -s http://localhost:8080/health 2>&1)

if echo "$HEALTH_CHECK" | grep -q "healthy"; then
    echo "✅ Backend API 健康狀態: 正常"
else
    echo "⚠️  Backend API 可能尚未完全啟動"
fi

echo ""
echo "=========================================="
echo "✅ PSM 系統啟動完成！"
echo ""
echo "訪問以下服務:"
echo "  🌐 Frontend:  http://localhost:3000"
echo "  🔌 Backend API: http://localhost:8080"
echo "  🗄️  Database:   localhost:5432"
echo ""
echo "查看日誌:"
echo "  docker-compose logs -f"
echo ""
echo "停止服務:"
echo "  docker-compose down"
echo ""
echo "=========================================="
