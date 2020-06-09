from deta.lib import app

@app.lib.run()
def runner(event):
    name = event.json.get('name', 'there')
    return f"Hello, sexy!"
