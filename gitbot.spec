%define COMMIT_HASH  %(cd $CI_PROJECT_DIR; git rev-parse --short HEAD 2>/dev/null)
%define BUILD_DATE   %(date -u +%Y%m%d%H%M)
%define GIT_MAJOR_VERSION %(cd $CI_PROJECT_DIR; git describe --abbrev=0 --tags| cut  -d'.' -f1)
%define GIT_MINOR_VERSION %(cd $CI_PROJECT_DIR; git describe --abbrev=0 --tags| cut  -d'.' -f2-3)


Name:           gitbot
Version:        %{GIT_MAJOR_VERSION}.%{GIT_MINOR_VERSION}
Release:        %{BUILD_DATE}%{?dist}
Summary:        Gitbot is a bot for gitlab  
Group:          Tools 
Packager:       Cosmin Ilie <cosmin_ilie@live.com>
License:        MIT.
Distribution:   . 
Vendor:         .
URL:            https://github.com/cosminilie/gitbot 
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
Source0:        gitbot.service
Source1:        gitbot.conf
Source2:	gitbot

%description
%{rpmname} is a bot for gitlab that updates permissions and configures a gitlab web hook. It aproves merge requests based on a couple of rules. 

%build
mkdir -p /tmp/gopath/{src,bin}/github.com/cosminilie
cp -r $CI_PROJECT_DIR  $GOPATH/src/github.com/cosminilie
go build -o /tmp/bin/%{name} -ldflags "-X main.majorVersion=${GIT_MAJOR_VERSION} -X main.minorVersion=${GIT_MINOR_VERSION} -X main.gitVersion=${COMMIT_HASH} -X main.buildDate=${BUILD_DATE}" $GOPATH/src/github.com/cosminilie/gitbot/cmd/main.go

%install
cp  -r $CI_PROJECT_DIR/build/* %{_sourcedir}
cp  /tmp/bin/%{name}  %{_sourcedir}

install -D -p -m 644 %{SOURCE0} %{buildroot}%{_unitdir}/%{name}.service
install -D -p -m 640 %{SOURCE1} %{buildroot}%{_sysconfdir}/%{name}/%{name}.conf
install -D -p -m 755 %{SOURCE2} %{buildroot}%{_bindir}/%{name}


%clean
rm -rf $RPM_BUILD_ROOT


%post
%systemd_post %{name}.service 

%preun
%systemd_preun %{name}.service 

%postun
%systemd_postun_with_restart %{name}.service 



%files
%defattr(644,root,root,755)
%{_unitdir}/%{name}.service
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.conf
%{_bindir}/%{name}



%doc
%changelog
