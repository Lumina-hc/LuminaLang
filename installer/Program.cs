using System;
using System.IO;
using System.Linq;
using System.Windows.Forms;

namespace LuminaInstaller
{
    internal static class Program
    {
        [STAThread]
        private static void Main(string[] args)
        {
            if (args.Length > 0 && (args[0] == "--silent" || args[0] == "-s"))
            {
                RunSilent(args);
                return;
            }

            ApplicationConfiguration.Initialize();
            Application.Run(new MainForm());
        }

        private static void RunSilent(string[] args)
        {
            string? dir = GetOption(args, "--dir") ?? Installer.DefaultDir();
            bool addToPath = args.Contains("--path");
            try
            {
                Installer.Install(dir, addToPath, Console.WriteLine);
                Console.WriteLine("installed: " + dir);
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine("install failed: " + ex.Message);
                Environment.Exit(1);
            }
        }

        private static string? GetOption(string[] args, string name)
        {
            for (int i = 0; i < args.Length; i++)
            {
                if (args[i] == name && i + 1 < args.Length)
                {
                    return args[i + 1];
                }
                if (args[i].StartsWith(name + "=", StringComparison.Ordinal))
                {
                    return args[i].Substring(name.Length + 1);
                }
            }
            return null;
        }
    }
}
