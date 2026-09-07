@echo off
title CIFO Enterprise Platform - Shutdown
echo ========================================================
echo  Menghentikan Seluruh Layanan CIFO Enterprise Platform
echo ========================================================
echo.
powershell -ExecutionPolicy Bypass -NoProfile -File "%~dp0scripts\stop-all.ps1"
echo.
echo Seluruh layanan telah dihentikan.
pause
