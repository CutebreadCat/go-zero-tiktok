#!/usr/bin/env bash
# log-cleaner.sh - 定时轮询清理 / 过滤服务日志
#
# 日志目录结构: logs/{service}/{date}/{service}.log 与 {service}_stderr.log
#   date 格式 YYYY-MM-DD,按天分目录。
#
# 功能:
#   1. 删除  超过保留天数的日期目录(默认保留 3 天)
#   2. 过滤  删除同目录下的 *_stderr.log,仅保留业务主日志
#
# 运行方式(守护模式,定时轮询):
#   nohup ./deploy/log-cleaner/log-cleaner.sh >/dev/null 2>&1 &
#   或写 systemd service(见文末注释)
#
# 可通过环境变量调整,例如:
#   LOG_ROOT=/app/tiktok/logs RETENTION=3 INTERVAL=3600 ./log-cleaner.sh
#
# 试运行(只看不删):
#   DRY_RUN=1 ./log-cleaner.sh

set -euo pipefail

# ---- 可配置项(环境变量覆盖) ----
LOG_ROOT="${LOG_ROOT:-logs}"          # 日志根目录
RETENTION="${RETENTION:-3}"           # 保留天数,超过即删除
INTERVAL="${INTERVAL:-3600}"          # 轮询间隔(秒),默认 1 小时
KEEP_STDERR="${KEEP_STDERR:-0}"       # 1=保留 _stderr.log,0=过滤删除
DRY_RUN="${DRY_RUN:-0}"               # 1=试运行,只打印不删除

# 保护:INTERVAL 必须为正整数,避免 sleep 0 忙循环
if ! [[ "$INTERVAL" =~ ^[0-9]+$ ]] || [ "$INTERVAL" -lt 1 ]; then
  echo "[警告] INTERVAL 非法('$INTERVAL'),回退为 3600" >&2
  INTERVAL=3600
fi

# ---- 单次清理 ----
clean_once() {
  # 保留截止日期:保留 RETENTION 天前及以后,更早的删除。
  # date 目录名 YYYY-MM-DD 字典序即时间序,可直接字符串比较。
  local cutoff deleted=0 filtered=0 day
  cutoff=$(date -d "-${RETENTION} days" +%F)

  # 遍历服务目录
  for svc_dir in "$LOG_ROOT"/*/; do
    [ -d "$svc_dir" ] || continue

    # 遍历日期目录
    for date_dir in "$svc_dir"*/; do
      [ -d "$date_dir" ] || continue
      day=$(basename "$date_dir")

      if [[ "$day" < "$cutoff" ]]; then
        # 超期:删除整个日期目录
        echo "[$(date '+%F %T')] 删除超期日志: $date_dir"
        if [ "$DRY_RUN" != 1 ]; then
          rm -rf "$date_dir"
        fi
        deleted=$((deleted + 1))
      else
        # 保留期内:过滤 *_stderr.log
        if [ "$KEEP_STDERR" != 1 ]; then
          for f in "$date_dir"*_stderr.log; do
            [ -f "$f" ] || continue
            echo "[$(date '+%F %T')] 过滤 stderr: $f"
            if [ "$DRY_RUN" != 1 ]; then
              rm -f "$f"
            fi
            filtered=$((filtered + 1))
          done
        fi
      fi
    done

    # 删除已清空的空服务目录
    if [ -z "$(ls -A "$svc_dir" 2>/dev/null)" ]; then
      echo "[$(date '+%F %T')] 删除空服务目录: $svc_dir"
      if [ "$DRY_RUN" != 1 ]; then
        rmdir "$svc_dir"
      fi
    fi
  done

  echo "[$(date '+%F %T')] log-cleaner 清理完成: 删除 $deleted 个日期目录, 过滤 $filtered 个 stderr (cutoff=$cutoff)"
}

# ---- 主流程:守护循环,定时轮询 ----
echo "[$(date '+%F %T')] log-cleaner 启动 (root=$LOG_ROOT, retention=${RETENTION}d, interval=${INTERVAL}s, dry_run=$DRY_RUN)"
while true; do
  clean_once
  sleep "$INTERVAL"
done

# systemd 托管示例 /etc/systemd/system/log-cleaner.service:
#   [Unit]
#   Description=Tiktok log cleaner
#   After=network.target
#   [Service]
#   Type=simple
#   WorkingDirectory=/opt/tiktok
#   Environment=LOG_ROOT=/opt/tiktok/logs
#   Environment=RETENTION=3
#   Environment=INTERVAL=3600
#   ExecStart=/opt/tiktok/deploy/log-cleaner/log-cleaner.sh
#   Restart=always
#   RestartSec=30
#   [Install]
#   WantedBy=multi-user.target