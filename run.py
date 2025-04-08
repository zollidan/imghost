import subprocess
import argparse
import os

def run_backend():
    print("🚀 Запуск Go (Fiber) backend...")
    subprocess.run(["go", "run", "main.go"], cwd="./backend")

def run_frontend():
    print("⚛️ Запуск React (Vite) frontend...")
    subprocess.run(["npm", "run", "dev"], cwd="./frontend", shell=True)

def main():
    parser = argparse.ArgumentParser(description="Run backend and/or frontend")

    parser.add_argument(
        "--backend", "-b",
        action="store_true",
        help="Запустить только backend (Go + Fiber)"
    )
    parser.add_argument(
        "--frontend", "-f",
        action="store_true",
        help="Запустить только frontend (React + Vite)"
    )

    args = parser.parse_args()

    # По умолчанию запускаем всё
    if not args.backend and not args.frontend:
        args.backend = args.frontend = True

    processes = []

    if args.backend:
        processes.append(subprocess.Popen(["go", "run", "main.go"], cwd="./backend"))

    if args.frontend:
        processes.append(subprocess.Popen(["npm", "run", "dev"], cwd="./frontend", shell=True))

    try:
        for p in processes:
            p.wait()
    except KeyboardInterrupt:
        print("⛔ Остановка...")
        for p in processes:
            p.terminate()

if __name__ == "__main__":
    main()
