from fastapi import FastAPI

app = FastAPI(title="Report Service", description="Python backend for report data")

@app.get("/ping")
def read_root():
    return {"message": "pong from Python report service"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8081, reload=True)
