@echo off
title CIFO Enterprise - Laptop 1 (Remote Mode)
echo ========================================================
echo  Menjalankan CIFO Workstation (Terkoneksi ke Laptop 2)
echo ========================================================
echo.
powershell -ExecutionPolicy Bypass -NoProfile -File "%~dp0scripts\start-remote.ps1"
echo.
echo Jendela launcher selesai. Layanan berjalan di jendela terpisah.
pause
