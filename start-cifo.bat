@echo off
title CIFO Enterprise Platform - Launcher
echo ========================================================
echo  Menjalankan CIFO Enterprise IT Monitoring ^& AIOps
echo ========================================================
echo.
powershell -ExecutionPolicy Bypass -NoProfile -File "%~dp0scripts\start-all.ps1"
echo.
echo Jendela launcher selesai. Layanan berjalan di jendela terpisah.
pause
