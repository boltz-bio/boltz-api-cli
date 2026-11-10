# Minimal org.freedesktop.secrets that models a LOCKED login keyring.
# Every Unlock returns a prompt. What the prompt does is read from /tmp/ss-mode:
#   hang    -> Prompt() returns but Completed is never emitted (nobody can answer
#              the unlock dialog: headless/SSH host). This is the default.
#   dismiss -> Completed(dismissed=True) is emitted (user/agent cancelled it).
import dbus, dbus.service, dbus.mainloop.glib
from gi.repository import GLib

SVC = "org.freedesktop.Secret.Service"
PROMPT = "org.freedesktop.Secret.Prompt"

def mode():
    try:
        return open("/tmp/ss-mode").read().strip()
    except OSError:
        return "hang"

class Prompt(dbus.service.Object):
    @dbus.service.method(PROMPT, in_signature="s")
    def Prompt(self, window_id):
        m = mode()
        print("prompt requested, mode=%s" % m, flush=True)
        if m == "dismiss":
            GLib.idle_add(lambda: self.Completed(True, dbus.String("", variant_level=1)) and False)

    @dbus.service.signal(PROMPT, signature="bv")
    def Completed(self, dismissed, result):
        pass

class Service(dbus.service.Object):
    n = 0
    @dbus.service.method(SVC, in_signature="ao", out_signature="aoo")
    def Unlock(self, objects):
        Service.n += 1
        path = "/org/freedesktop/secrets/prompt/p%d" % Service.n
        Prompt(self.connection, path)
        print("Unlock(%s) -> prompt %s" % (list(objects), path), flush=True)
        return dbus.Array([], signature="o"), dbus.ObjectPath(path)

    @dbus.service.method(SVC, in_signature="sv", out_signature="vo")
    def OpenSession(self, alg, inp):
        return dbus.String("", variant_level=1), dbus.ObjectPath("/org/freedesktop/secrets/session/s1")

    @dbus.service.method(dbus.PROPERTIES_IFACE, in_signature="ss", out_signature="v")
    def Get(self, iface, prop):
        return dbus.Array([], signature="o", variant_level=1)

dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)
bus = dbus.SessionBus()
name = dbus.service.BusName("org.freedesktop.secrets", bus)
Service(bus, "/org/freedesktop/secrets")
print("fake secret service ready", flush=True)
GLib.MainLoop().run()
