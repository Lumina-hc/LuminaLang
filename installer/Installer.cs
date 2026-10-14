using System;
using System.IO;
using System.Linq;
using System.Reflection;
using System.Runtime.InteropServices;
using Microsoft.Win32;

namespace LuminaInstaller
{
    internal static class Installer
    {
        public const string AppVersion = "1.0.0";

        public static string DefaultDir()
        {
            string baseDir = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
            return Path.Combine(baseDir, "Programs", "LuminaLang");
        }

        public static void Install(string dir, bool addToPath, Action<string>? report)
        {
            report?.Invoke("正在创建安装目录 ...");
            Directory.CreateDirectory(dir);

            report?.Invoke("正在写入 luminalang.exe ...");
            Extract("luminalang.exe", Path.Combine(dir, "luminalang.exe"));

            report?.Invoke("正在写入 pho.exe ...");
            Extract("pho.exe", Path.Combine(dir, "pho.exe"));

            if (addToPath)
            {
                report?.Invoke("正在更新 PATH ...");
                AddToUserPath(dir);
            }

            report?.Invoke("安装完成。");
        }

        private static void Extract(string resourceName, string destination)
        {
            var assembly = Assembly.GetExecutingAssembly();
            using Stream source = assembly.GetManifestResourceStream(resourceName)
                ?? throw new InvalidOperationException("缺少内嵌资源：" + resourceName);
            using FileStream target = File.Create(destination);
            source.CopyTo(target);
        }

        private static void AddToUserPath(string dir)
        {
            using RegistryKey key = Registry.CurrentUser.OpenSubKey("Environment", writable: true)
                ?? Registry.CurrentUser.CreateSubKey("Environment")
                ?? throw new InvalidOperationException("无法打开环境变量注册表项。");

            string current = key.GetValue("Path", "", RegistryValueOptions.DoNotExpandEnvironmentNames) as string ?? "";
            string[] parts = current.Split(';', StringSplitOptions.RemoveEmptyEntries);
            string target = dir.TrimEnd('\\');

            if (parts.Any(p => string.Equals(p.TrimEnd('\\'), target, StringComparison.OrdinalIgnoreCase)))
            {
                return;
            }

            string updated = parts.Length == 0 ? dir : string.Join(";", parts) + ";" + dir;
            key.SetValue("Path", updated, RegistryValueKind.ExpandString);
            BroadcastChange();
        }

        [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
        private static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint msg, UIntPtr wParam, string lParam, uint flags, uint timeout, out UIntPtr result);

        private static void BroadcastChange()
        {
            SendMessageTimeout(new IntPtr(0xffff), 0x001A, UIntPtr.Zero, "Environment", 0x0002, 5000, out _);
        }
    }
}
